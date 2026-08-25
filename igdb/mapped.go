package igdb

type Platform struct {
	ID       PlatformID // According to IGDB
	Slug     string     // Widely understood short name, eg "NES" or "SNES"
	Includes []int
	Names    []string
	Suffix   []string // Dot product with Names, eg "Family Computer" and "Disc System"
}

var MappedPlatforms = map[string]map[string][]Platform{
	"Nintendo": {
		"NES/Famicom": {
			{18, "NES", []int{99}, []string{"Nintendo Entertainment System"}, []string{}},
			{99, "FC", []int{18}, []string{"Famicom", "Family Computer"}, []string{}},
			{51, "FC-Disc", []int{99}, []string{"Famicom", "Family Computer"}, []string{"Disc System"}},
		},
		"SNES/Super Famicom": {
			{19, "SNES", []int{58}, []string{"Super Nintendo Entertainment System", "Super NES"}, []string{}},
			{58, "SFC", []int{19}, []string{"Super Famicom", "Super Family Computer"}, []string{}},
			{131, "SFC-CD", []int{58}, []string{"Super Famicom", "Super Family Computer"}, []string{"CD-ROM System"}},
			{306, "SFC-TV", []int{58}, []string{"Satellaview", "Super Famicom TV", "Super Family Computer TV"}, []string{}},
		},
		"GameBoy": {
			{33, "GB", []int{}, []string{"GameBoy", "Game Boy"}, []string{}},
			{22, "GBC", []int{33}, []string{"GameBoy", "Game Boy"}, []string{"Color"}},
			{24, "GBA", []int{22, 33}, []string{"GameBoy", "Game Boy"}, []string{"Advance"}},
			{510, "GBCard", []int{}, []string{"e-Reader", "Card-e Reader"}, []string{}},
			{166, "PKMini", []int{}, []string{"Pokemon Mini", "Pokémon mini"}, []string{}},
		},
		"DS": {
			{20, "DS", []int{}, []string{"Nintendo DS"}, []string{}},
			{159, "DSi", []int{20}, []string{"Nintendo DSi"}, []string{}},
			{37, "3DS", []int{20, 159}, []string{"Nintendo 3DS", "2DS", "Nintendo 2DS"}, []string{}},
			{137, "New3DS", []int{37, 159, 20}, []string{"New Nintendo 3DS", "New 3DS"}, []string{}},
		},
		"Wii": {
			{5, "Wii", []int{}, []string{}, []string{}},
			{41, "WiiU", []int{5}, []string{"Wii U"}, []string{}},
			{47, "VC", []int{}, []string{"Virtual Console"}, []string{}},
		},
		"Nintendo 64": {
			{4, "N64", []int{}, []string{"Nintendo 64"}, []string{}},
			{416, "64DD", []int{4}, []string{"N64", "Nintendo 64"}, []string{"DD", "Disc System"}},
		},
		"Switch": {
			{130, "NS1", []int{}, []string{"Nintendo Switch", "Switch"}, []string{}},
			{508, "NS2", []int{130}, []string{"Nintendo Switch", "Switch"}, []string{"2"}},
		},
		"GameCube": {
			{21, "GCN", []int{}, []string{"GameCube", "Game Cube"}, []string{}},
		},
		"Game & Watch": {
			{307, "G&W", []int{}, []string{"Game & Watch", "Game and Watch", "Game n Watch"}, []string{}},
		},
		"Virtual Boy": {
			{87, "VBN", []int{}, []string{"Virtual Boy", "VirtualBoy"}, []string{}},
		},
	},
	"Home Computer (Retro)": {
		"Commodore": {
			{15, "C64", []int{}, []string{"Commodore"}, []string{"64", "C64", "128", "C128", "MAX"}},
			{16, "Amiga", []int{}, []string{"Commodore Amiga", "Amiga"}, []string{}},
			{71, "VIC-20", []int{}, []string{"Commodore"}, []string{"VIC-20", "VIC20", "VC"}},
			{93, "C16", []int{94}, []string{"Commodore"}, []string{"16", "C16"}},
			{94, "CP4", []int{93}, []string{"Commodore"}, []string{"Plus 4", "Plus/4", "+4", "Plus4"}},
			{90, "PET", []int{}, []string{"Commodore"}, []string{"PET"}},
			{158, "CDTV", []int{}, []string{"Commodore"}, []string{"CDTV", "CD TV"}},
			{114, "AmigaCD32", []int{16}, []string{"Commodore Amiga", "Amiga"}, []string{"CD32", "CD 32"}},
		},
		"Acorn": {
			{116, "Archimedes", []int{}, []string{"Acorn Archimedes"}, []string{}},
			{69, "BBC", []int{134}, []string{"BBC Microcomputer System"}, []string{}},
			{134, "Electron", []int{69}, []string{"Acorn Electron"}, []string{}},
		},
		"Amstrad": {
			{25, "CPC", []int{}, []string{"Amstrad"}, []string{"CPC", "464"}},
			{154, "PCW", []int{}, []string{"Amstrad"}, []string{"PCW"}},
		},
		"Apple": {
			{75, "A2", []int{}, []string{"Apple"}, []string{"II", "][", "2"}},
			{115, "A2gs", []int{75}, []string{"Apple"}, []string{"IIGS", "II GS", "2GS", "2 GS"}},
		},
	},
	"Virtual Reality": {
		"Valve": {
			{163, "SteamVR", []int{}, []string{"Steam", "Valve"}, []string{"VR"}},
		},
		"Meta/Oculus": {
			{387, "O_Go", []int{}, []string{"Oculus"}, []string{"Go"}},
			{385, "O_Rift", []int{}, []string{"Oculus"}, []string{"Rift"}},
			{384, "O_Quest1", []int{385}, []string{"Oculus"}, []string{"Quest", "1"}},
			{386, "O_Quest2", []int{384, 385}, []string{"Meta", "Oculus", "Meta Oculus"}, []string{"Quest 2", "2"}},
			{471, "O_Quest3", []int{386, 385, 384}, []string{"Meta", "Oculus", "Meta Oculus"}, []string{"Quest 3", "3"}},
		},
	},
	"Phone/Mobile": {
		"Apple": {
			{39, "iOS", []int{}, []string{"Apple iOS", "iOS"}, []string{}},
		},
		"Android": {
			{34, "Android", []int{132}, []string{"Android (Generic)"}, []string{}},
			{132, "FireTV", []int{}, []string{"Amazon"}, []string{"FireTV", "Fire TV"}},
		},
	},
	"Arcade": {
		"Neo Geo": {
			{79, "NG_MVS", []int{}, []string{"Neo Geo", "NeoGeo"}, []string{"MVS"}},
			{80, "NG_AES", []int{}, []string{"Neo Geo", "NeoGeo"}, []string{"AES"}},
			{136, "NG_CD", []int{}, []string{"Neo Geo", "NeoGeo"}, []string{"CD"}},
			{135, "NG_64", []int{}, []string{"Hyper Neo Geo", "Hyper NeoGeo"}, []string{"64"}},
			{119, "NG_P", []int{}, []string{"Neo Geo", "NeoGeo"}, []string{"Pocket"}},
			{120, "NG_PC", []int{119}, []string{"Neo Geo", "NeoGeo"}, []string{"Pocket Color"}},
		},
	},
}
