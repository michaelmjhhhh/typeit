#!/usr/bin/env python3
"""Exercise a real terminal, three stages, persistence, history, and restoration."""
from pathlib import Path
import argparse, errno, fcntl, json, os, pty, select, signal, sqlite3, struct, tempfile, termios, time
options=argparse.ArgumentParser(description=__doc__);options.add_argument('--record',type=Path);args=options.parse_args()
root=Path(__file__).resolve().parent.parent
binary=root/'bin/typeit'
code='package main\nfunc main() {\n    println("hello")\n}\n'
expected='package main\nfunc main() {\nprintln("hello")\n}'
class Terminal:
 def __init__(self,args,data):
  self.output=b'';self.cursor=0;self.events=[];self.started=time.monotonic()
  self.pid,self.fd=pty.fork()
  if self.pid==0:
   os.environ.update(TERM='xterm-256color',GITTYPE_DATA_DIR=str(data))
   os.execv(str(binary),[str(binary)]+args)
  fcntl.ioctl(self.fd,termios.TIOCSWINSZ,struct.pack('HHHH',32,110,0,0))
 def read(self,seconds=.1):
  deadline=time.monotonic()+seconds
  while time.monotonic()<deadline:
   ready,_,_=select.select([self.fd],[],[],max(0,deadline-time.monotonic()))
   if not ready:break
   try:b=os.read(self.fd,65536)
   except OSError as e:
    if e.errno==errno.EIO:break
    raise
   if not b:break
   self.output+=b;self.events.append([round(time.monotonic()-self.started,6),"o",b.decode("utf-8",errors="replace")])
 def wait(self,text,timeout=15):
  deadline=time.monotonic()+timeout;needle=text.encode()
  while time.monotonic()<deadline:
   self.read(.1)
   position=self.output.find(needle,self.cursor)
   if position>=0:self.cursor=position+len(needle);return
  raise AssertionError(f'terminal did not display {text!r}: {self.output[-2500:]!r}')
 def send(self,text):os.write(self.fd,text.encode());self.read(.1)
 def finish(self):
  self.send('\x03');self.read(1)
  pid,status=os.waitpid(self.pid,os.WNOHANG)
  if not pid:
   os.kill(self.pid,signal.SIGTERM);os.waitpid(self.pid,0);raise AssertionError('terminal did not exit')
  if not os.WIFEXITED(status) or os.WEXITSTATUS(status)!=0:raise AssertionError(f'exit status {status}')
  if b'\x1b[?1049l' not in self.output:raise AssertionError('alternate screen was not restored')
  if b'\x1b[?25h' not in self.output:raise AssertionError('cursor was not restored')
  os.close(self.fd)
with tempfile.TemporaryDirectory(prefix='gittype-smoke-') as temp:
 temp=Path(temp);source=temp/'source';source.mkdir();(source/'main.go').write_text(code);data=temp/'data'
 terminal=Terminal([str(source),'--langs','go'],data)
 try:
  terminal.wait('Code Typing Challenge');terminal.send('lll');terminal.send(' ');terminal.wait('Press SPACE to start')
  for stage in range(3):
   terminal.send(' ');terminal.read(2.6)
   for char in expected:terminal.send('\r' if char=='\n' else char)
   terminal.wait('STAGE COMPLETE');terminal.send(' ')
   if stage<2:terminal.wait('Press SPACE to start')
  terminal.wait('ANALYZING PERFORMANCE');terminal.read(.4);terminal.send('s');terminal.wait('SESSION COMPLETE')
  terminal.send('d');terminal.wait('SESSION DETAILS');terminal.send('\x1b');terminal.read(.3);terminal.send('t');terminal.wait('Code Typing Challenge');terminal.send('r');terminal.wait('SESSION HISTORY');terminal.send('\r');terminal.wait('SESSION DETAILS')
  terminal.send('\x1b');terminal.read(.3);terminal.send('\x1b');terminal.wait('Code Typing Challenge')
  terminal.send('a');terminal.wait('ANALYTICS')
  for tab in range(3):terminal.send('\t')
  terminal.send('\x1b');terminal.wait('Code Typing Challenge');terminal.send('s');terminal.wait('SETTINGS');terminal.send('\x1b[B');terminal.send('\x1b[C');terminal.send('\x1b[B');terminal.send('\x1b');terminal.wait('Code Typing Challenge')
  terminal.finish()
  if args.record:
   args.record.parent.mkdir(parents=True,exist_ok=True)
   with args.record.open('w') as recording:
    recording.write(json.dumps({'version':2,'width':110,'height':32,'title':'GitType Go: three-stage session, history, analytics, settings'})+'\n')
    for event in terminal.events:recording.write(json.dumps(event)+'\n')
 except BaseException:
  try:os.kill(terminal.pid,signal.SIGKILL);os.waitpid(terminal.pid,0)
  except ProcessLookupError:pass
  raise
 db=sqlite3.connect(data/'gittype.db')
 row=db.execute('SELECT stages_completed,stages_attempted,mistakes,accuracy FROM session_results').fetchone()
 assert row==(3,3,0,100.0),row
 assert db.execute('SELECT COUNT(*) FROM stage_results').fetchone()[0]==3
 print(json.dumps({'terminal':'110x32','completed_stages':3,'mistakes':0,'accuracy':100,'history_details':True,'terminal_restored':True},indent=2))
