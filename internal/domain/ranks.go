package domain

import "math"

type Rank struct {
	Name, Tier string
	Min, Max   uint32
}

var Ranks = []Rank{
	{"Hello World", "Beginner", 0, 800},
	{"Syntax Error", "Beginner", 801, 1200},
	{"Rubber Duck", "Beginner", 1201, 1600},
	{"Script Kid", "Beginner", 1601, 2000},
	{"Bash Newbie", "Beginner", 2001, 2450},
	{"CLI Wanderer", "Beginner", 2451, 2900},
	{"Tab Tamer", "Beginner", 2901, 3300},
	{"Bracket Juggler", "Beginner", 3301, 3700},
	{"Copy-Paste Engineer", "Beginner", 3701, 4150},
	{"Linter Apprentice", "Beginner", 4151, 4550},
	{"Unit Test Trainee", "Beginner", 4551, 5000},
	{"Code Monkey", "Beginner", 5001, 5600},
	{"Ticket Picker", "Intermediate", 5601, 5850},
	{"Junior Dev", "Intermediate", 5851, 6000},
	{"Git Ninja", "Intermediate", 6001, 6100},
	{"Merge Wrangler", "Intermediate", 6101, 6250},
	{"API Crafter", "Intermediate", 6251, 6400},
	{"Frontend Dev", "Intermediate", 6401, 6550},
	{"Backend Dev", "Intermediate", 6551, 6700},
	{"CI Tinkerer", "Intermediate", 6701, 6850},
	{"Test Pilot", "Intermediate", 6851, 7000},
	{"Build Tamer", "Intermediate", 7001, 7100},
	{"Code Reviewer", "Intermediate", 7101, 7250},
	{"Release Handler", "Intermediate", 7251, 7500},
	{"Refactorer", "Advanced", 7501, 7800},
	{"Senior Dev", "Advanced", 7801, 8000},
	{"DevOps Engineer", "Advanced", 8001, 8100},
	{"Incident Responder", "Advanced", 8101, 8250},
	{"Reliability Guardian", "Advanced", 8251, 8400},
	{"Security Engineer", "Advanced", 8401, 8500},
	{"Performance Alchemist", "Advanced", 8501, 8650},
	{"Data Pipeline Master", "Advanced", 8651, 8800},
	{"Tech Lead", "Advanced", 8801, 8950},
	{"Architect", "Advanced", 8951, 9100},
	{"Protocol Artisan", "Advanced", 9101, 9250},
	{"Kernel Hacker", "Advanced", 9251, 9500},
	{"Compiler", "Expert", 9501, 9800},
	{"Bytecode Interpreter", "Expert", 9801, 9950},
	{"Virtual Machine", "Expert", 9951, 10100},
	{"Operating System", "Expert", 10101, 10200},
	{"Filesystem", "Expert", 10201, 10350},
	{"Network Stack", "Expert", 10351, 10500},
	{"Database Engine", "Expert", 10501, 10650},
	{"Query Optimizer", "Expert", 10651, 10800},
	{"Cloud Platform", "Expert", 10801, 10950},
	{"Container Orchestrator", "Expert", 10951, 11100},
	{"Stream Processor", "Expert", 11101, 11200},
	{"Quantum Computer", "Expert", 11201, 11400},
	{"GPU Cluster", "Legendary", 11401, 11700},
	{"DNS Overlord", "Legendary", 11701, 12250},
	{"CDN Sentinel", "Legendary", 12251, 12800},
	{"Load Balancer Primarch", "Legendary", 12801, 13400},
	{"Singularity", "Legendary", 13401, 13950},
	{"The Machine", "Legendary", 13951, 14500},
	{"Origin", "Legendary", 14501, 15100},
	{"SegFault", "Legendary", 15101, 15650},
	{"Buffer Overflow", "Legendary", 15651, 16200},
	{"Memory Leak", "Legendary", 16201, 16800},
	{"Null Pointer Exception", "Legendary", 16801, 17350},
	{"Undefined Behavior", "Legendary", 17351, 17900},
	{"Heisenbug", "Legendary", 17901, 18500},
	{"Blue Screen", "Legendary", 18501, 19100},
	{"Kernel Panic", "Legendary", 19101, 4294967295},
}

func RankFor(score float64) Rank {
	if math.IsNaN(score) || score < 0 {
		score = 0
	}
	for _, r := range Ranks {
		if score < float64(r.Max)+1 {
			return r
		}
	}
	return Ranks[len(Ranks)-1]
}
