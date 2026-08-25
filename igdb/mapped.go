package igdb

type Platform struct {
	ID       PlatformID // According to IGDB
	Slug     string     // Widely understood short name, eg "NES" or "SNES"
	Includes []int
	Names    []string
	Suffix   []string // Dot product with Names, eg "Family Computer" and "Disc System"
}

// MappedPlatforms is the picker tree: manufacturer → family → consoles.
// Comment a matching entry in platforms.go once it lives here.
// DUPLICATE Stadia (203) is left unmapped on purpose.
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
	"PlayStation": {
		"Home": {
			{7, "PS1", []int{}, []string{"PlayStation", "PS1", "PSX"}, []string{}},
			{8, "PS2", []int{7}, []string{"PlayStation"}, []string{"2", "PS2"}},
			{9, "PS3", []int{8, 7}, []string{"PlayStation"}, []string{"3", "PS3"}},
			{48, "PS4", []int{9}, []string{"PlayStation"}, []string{"4", "PS4"}},
			{167, "PS5", []int{48, 9}, []string{"PlayStation"}, []string{"5", "PS5"}},
		},
		"Portable": {
			{38, "PSP", []int{}, []string{"PlayStation Portable", "PSP"}, []string{}},
			{46, "PSV", []int{38}, []string{"PlayStation Vita", "PS Vita", "PSV"}, []string{}},
			{441, "PKST", []int{7}, []string{"PocketStation"}, []string{}},
		},
		"VR": {
			{165, "PSVR", []int{48}, []string{"PlayStation VR", "PS VR"}, []string{}},
			{390, "PSVR2", []int{165, 167}, []string{"PlayStation VR2", "PS VR2"}, []string{}},
		},
	},
	"Xbox": {
		"Xbox": {
			{11, "XB", []int{}, []string{"Xbox"}, []string{}},
			{12, "X360", []int{11}, []string{"Xbox"}, []string{"360"}},
			{49, "XB1", []int{12}, []string{"Xbox One", "Xbox"}, []string{"One"}},
			{169, "XSX", []int{49, 12}, []string{"Xbox Series X|S", "Xbox Series", "Series X", "Series S"}, []string{}},
		},
	},
	"PC": {
		"Windows": {
			{6, "Win", []int{13}, []string{"Windows", "PC (Microsoft Windows)", "PC"}, []string{}},
		},
		"macOS": {
			{14, "Mac", []int{}, []string{"macOS", "Mac", "Macintosh", "OS X"}, []string{}},
		},
		"Linux": {
			{3, "Linux", []int{}, []string{"Linux"}, []string{}},
		},
		"Legacy": {
			{13, "DOS", []int{}, []string{"DOS", "MS-DOS", "PC DOS"}, []string{}},
			{409, "LegacyPC", []int{}, []string{"Legacy Computer"}, []string{}},
			{77, "X1", []int{}, []string{"Sharp"}, []string{"X1"}},
			{121, "X68k", []int{77}, []string{"Sharp"}, []string{"X68000", "X68k", "X6800"}},
			{374, "MZ2200", []int{}, []string{"Sharp"}, []string{"MZ-2200", "MZ2200"}},
			{149, "PC98", []int{}, []string{"NEC PC-9800", "PC-98", "PC98"}, []string{}},
			{118, "FMT", []int{}, []string{"FM Towns", "Fujitsu FM Towns"}, []string{}},
			{236, "Sorcerer", []int{}, []string{"Exidy Sorcerer"}, []string{}},
			{237, "Sol20", []int{}, []string{"Sol-20", "Processor Technology Sol-20"}, []string{}},
		},
	},
	"Home Computer": {
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
		"Sinclair": {
			{26, "ZX", []int{}, []string{"ZX Spectrum", "Sinclair ZX Spectrum", "Spectrum"}, []string{}},
			{373, "ZX81", []int{}, []string{"Sinclair"}, []string{"ZX81", "ZX-81", "ZX 81"}},
			{406, "QL", []int{}, []string{"Sinclair"}, []string{"QL"}},
		},
		"Acorn": {
			{116, "Archimedes", []int{}, []string{"Acorn Archimedes"}, []string{}},
			{69, "BBC", []int{134}, []string{"BBC Microcomputer System", "BBC Micro"}, []string{}},
			{134, "Electron", []int{69}, []string{"Acorn Electron"}, []string{}},
		},
		"Amstrad": {
			{25, "CPC", []int{}, []string{"Amstrad"}, []string{"CPC", "464"}},
			{154, "PCW", []int{}, []string{"Amstrad"}, []string{"PCW"}},
			{506, "GX4000", []int{25}, []string{"Amstrad"}, []string{"GX4000", "GX 4000"}},
		},
		"Apple": {
			{75, "A2", []int{}, []string{"Apple"}, []string{"II", "][", "2"}},
			{115, "A2gs", []int{75}, []string{"Apple"}, []string{"IIGS", "II GS", "2GS", "2 GS"}},
			{476, "Pippin", []int{}, []string{"Apple Pippin", "Pippin", "Bandai Pippin"}, []string{}},
		},
		"MSX": {
			{27, "MSX", []int{}, []string{"MSX"}, []string{}},
			{53, "MSX2", []int{27}, []string{"MSX"}, []string{"2", "MSX2"}},
		},
		"NEC": {
			{125, "PC88", []int{}, []string{"NEC PC-8800", "PC-88", "PC88"}, []string{}},
			{157, "PC60", []int{}, []string{"NEC PC-6000", "PC-6001", "PC-60"}, []string{}},
		},
		"Fujitsu": {
			{152, "FM7", []int{}, []string{"FM-7", "Fujitsu FM-7"}, []string{}},
		},
		"Tandy": {
			{126, "TRS80", []int{}, []string{"TRS-80", "Tandy TRS-80"}, []string{}},
			{151, "CoCo", []int{126}, []string{"TRS-80 Color Computer", "Tandy Color Computer", "CoCo"}, []string{}},
		},
		"Texas Instruments": {
			{129, "TI99", []int{}, []string{"Texas Instruments TI-99", "TI-99", "TI-99/4A"}, []string{}},
		},
		"Dragon": {
			{153, "Dragon", []int{}, []string{"Dragon 32/64", "Dragon 32", "Dragon 64"}, []string{}},
		},
		"Thomson": {
			{156, "MO5", []int{}, []string{"Thomson MO5"}, []string{}},
		},
		"Tatung": {
			{155, "Einstein", []int{}, []string{"Tatung Einstein"}, []string{}},
		},
		"Tomy": {
			{481, "Tutor", []int{}, []string{"Tomy Tutor", "Pyuta", "Grandstand Tutor"}, []string{}},
		},
	},
	"Retro": {
		"Atari": {
			{59, "A2600", []int{}, []string{"Atari"}, []string{"2600", "VCS"}},
			{66, "A5200", []int{59}, []string{"Atari"}, []string{"5200"}},
			{60, "A7800", []int{59}, []string{"Atari"}, []string{"7800"}},
			{65, "A8", []int{}, []string{"Atari"}, []string{"8-bit", "400", "800", "XL", "XE"}},
			{63, "ST", []int{}, []string{"Atari"}, []string{"ST", "STE", "ST/STE"}},
			{61, "Lynx", []int{}, []string{"Atari Lynx"}, []string{}},
			{62, "Jaguar", []int{}, []string{"Atari Jaguar"}, []string{}},
			{410, "JaguarCD", []int{62}, []string{"Atari Jaguar"}, []string{"CD"}},
		},
		"Sega": {
			{84, "SG1K", []int{}, []string{"SG-1000", "Sega SG-1000"}, []string{}},
			{64, "SMS", []int{84}, []string{"Sega Master System", "Master System", "Mark III"}, []string{}},
			{29, "GEN", []int{}, []string{"Sega Mega Drive", "Sega Genesis", "Mega Drive", "Genesis"}, []string{}},
			{30, "32X", []int{29}, []string{"Sega"}, []string{"32X", "Super 32X"}},
			{78, "SCD", []int{29}, []string{"Sega CD", "Mega-CD", "Sega Mega-CD"}, []string{}},
			{482, "SCD32X", []int{78, 30, 29}, []string{"Sega CD 32X", "Sega Mega-CD 32X"}, []string{}},
			{32, "SAT", []int{}, []string{"Sega Saturn", "Saturn"}, []string{}},
			{23, "DC", []int{}, []string{"Dreamcast", "Sega Dreamcast"}, []string{}},
			{35, "GG", []int{}, []string{"Sega Game Gear", "Game Gear"}, []string{}},
			{339, "Pico", []int{}, []string{"Sega Pico", "Pico"}, []string{}},
			{507, "Beena", []int{339}, []string{"Advanced Pico Beena", "Beena"}, []string{}},
			{440, "VMU", []int{23}, []string{"Visual Memory Unit", "Visual Memory System", "VMU", "VMS"}, []string{}},
		},
		"NEC": {
			{86, "TG16", []int{}, []string{"TurboGrafx-16", "PC Engine", "TurboGrafx-16/PC Engine"}, []string{}},
			{128, "SGX", []int{86}, []string{"PC Engine SuperGrafx", "SuperGrafx"}, []string{}},
			{150, "TGCD", []int{86}, []string{"TurboGrafx-16/PC Engine CD", "TurboGrafx-CD", "PC Engine CD-ROM²"}, []string{}},
			{274, "PCFX", []int{}, []string{"PC-FX", "NEC PC-FX"}, []string{}},
		},
		"Coleco": {
			{68, "Coleco", []int{}, []string{"ColecoVision"}, []string{}},
		},
		"Mattel": {
			{67, "INTV", []int{}, []string{"Intellivision"}, []string{}},
			{382, "Amico", []int{67}, []string{"Intellivision Amico"}, []string{}},
			{407, "HyperScan", []int{}, []string{"HyperScan", "Mattel HyperScan"}, []string{}},
		},
		"Magnavox": {
			{88, "Odyssey", []int{}, []string{"Odyssey", "Magnavox Odyssey"}, []string{}},
			{133, "Odyssey2", []int{88}, []string{"Odyssey 2", "Videopac G7000", "Odyssey 2 / Videopac G7000"}, []string{}},
		},
		"Fairchild": {
			{127, "ChannelF", []int{}, []string{"Fairchild Channel F", "Channel F"}, []string{}},
		},
		"GCE": {
			{70, "Vectrex", []int{}, []string{"Vectrex"}, []string{}},
		},
		"3DO": {
			{50, "3DO", []int{}, []string{"3DO Interactive Multiplayer", "3DO"}, []string{}},
		},
		"Bandai": {
			{57, "WS", []int{}, []string{"WonderSwan"}, []string{}},
			{123, "WSC", []int{57}, []string{"WonderSwan"}, []string{"Color"}},
			{124, "SwanX", []int{123, 57}, []string{"SwanCrystal"}, []string{}},
			{308, "Playdia", []int{}, []string{"Playdia"}, []string{}},
		},
		"Epoch": {
			{375, "Cassette", []int{}, []string{"Epoch Cassette Vision"}, []string{}},
			{376, "SuperCV", []int{375}, []string{"Epoch Super Cassette Vision"}, []string{}},
		},
		"Philips": {
			{117, "CDi", []int{}, []string{"Philips CD-i", "CD-i"}, []string{}},
		},
		"Bally": {
			{91, "Astrocade", []int{}, []string{"Bally Astrocade", "Astrocade"}, []string{}},
		},
		"Milton Bradley": {
			{89, "Microvision", []int{}, []string{"Microvision"}, []string{}},
		},
		"Emerson": {
			{473, "Arcadia", []int{}, []string{"Arcadia 2001"}, []string{}},
		},
		"Interton": {
			{138, "VC4000", []int{}, []string{"VC 4000", "Interton VC 4000"}, []string{}},
			{139, "1292", []int{}, []string{"1292 Advanced Programmable Video System"}, []string{}},
		},
		"Tiger": {
			{379, "Gamecom", []int{}, []string{"Game.com", "Tiger Game.com"}, []string{}},
			{475, "RZone", []int{}, []string{"R-Zone", "Tiger R-Zone"}, []string{}},
		},
		"Watara": {
			{415, "Supervision", []int{}, []string{"Watara/QuickShot Supervision", "Watara Supervision", "Supervision"}, []string{}},
		},
		"Casio": {
			{380, "Loopy", []int{}, []string{"Casio Loopy"}, []string{}},
		},
		"Panasonic": {
			{477, "Jungle", []int{}, []string{"Panasonic Jungle"}, []string{}},
			{478, "M2", []int{}, []string{"Panasonic M2"}, []string{}},
		},
		"Pioneer": {
			{487, "LaserActive", []int{}, []string{"LaserActive", "Pioneer LaserActive"}, []string{}},
		},
		"VM Labs": {
			{122, "Nuon", []int{}, []string{"Nuon"}, []string{}},
		},
		"Funtech": {
			{480, "SuperACan", []int{}, []string{"Super A'Can"}, []string{}},
		},
		"Bit Corp": {
			{378, "Gamate", []int{}, []string{"Gamate"}, []string{}},
		},
		"Welback": {
			{408, "MegaDuck", []int{}, []string{"Mega Duck", "Cougar Boy", "Mega Duck/Cougar Boy"}, []string{}},
		},
		"Blaze": {
			{309, "Evercade", []int{}, []string{"Evercade"}, []string{}},
		},
		"Playmaji": {
			{509, "Polymega", []int{}, []string{"Polymega"}, []string{}},
		},
		"Digital Blue": {
			{486, "Digiblast", []int{}, []string{"Digiblast"}, []string{}},
		},
		"Dedicated": {
			{377, "PnP", []int{}, []string{"Plug & Play"}, []string{}},
			{411, "LCD", []int{}, []string{"Handheld Electronic LCD"}, []string{}},
			{142, "PC50X", []int{}, []string{"PC-50X Family", "PC-50X"}, []string{}},
		},
	},
	"Virtual Reality": {
		"Valve": {
			{163, "SteamVR", []int{}, []string{"Steam", "Valve"}, []string{"VR"}},
		},
		"Meta/Oculus": {
			{162, "Oculus", []int{}, []string{"Oculus VR", "Oculus"}, []string{}},
			{387, "O_Go", []int{}, []string{"Oculus"}, []string{"Go"}},
			{385, "O_Rift", []int{}, []string{"Oculus"}, []string{"Rift"}},
			{384, "O_Quest1", []int{385}, []string{"Oculus"}, []string{"Quest", "1"}},
			{386, "O_Quest2", []int{384, 385}, []string{"Meta", "Oculus", "Meta Oculus"}, []string{"Quest 2", "2"}},
			{471, "O_Quest3", []int{386, 385, 384}, []string{"Meta", "Oculus", "Meta Oculus"}, []string{"Quest 3", "3"}},
		},
		"PlayStation": {
			{165, "PSVR", []int{48}, []string{"PlayStation VR", "PS VR"}, []string{}},
			{390, "PSVR2", []int{165, 167}, []string{"PlayStation VR2", "PS VR2"}, []string{}},
		},
		"Microsoft": {
			{161, "WMR", []int{}, []string{"Windows Mixed Reality"}, []string{}},
		},
		"Google": {
			{164, "Daydream", []int{}, []string{"Daydream", "Google Daydream"}, []string{}},
		},
		"Samsung": {
			{388, "GearVR", []int{}, []string{"Gear VR", "Samsung Gear VR"}, []string{}},
		},
		"Apple": {
			{472, "visionOS", []int{}, []string{"visionOS", "Apple Vision", "Vision Pro"}, []string{}},
		},
	},
	"Phone/Mobile": {
		"Apple": {
			{39, "iOS", []int{}, []string{"Apple iOS", "iOS"}, []string{}},
		},
		"Android": {
			{34, "Android", []int{132}, []string{"Android (Generic)", "Android"}, []string{}},
			{132, "FireTV", []int{}, []string{"Amazon"}, []string{"FireTV", "Fire TV"}},
			{72, "Ouya", []int{34}, []string{"Ouya"}, []string{}},
			{240, "Zeebo", []int{}, []string{"Zeebo"}, []string{}},
		},
		"Nokia": {
			{42, "NGage", []int{}, []string{"N-Gage"}, []string{}},
		},
		"Microsoft": {
			{74, "WP", []int{}, []string{"Windows Phone"}, []string{}},
			{405, "WinMob", []int{}, []string{"Windows Mobile"}, []string{}},
		},
		"BlackBerry": {
			{73, "BB", []int{}, []string{"BlackBerry OS", "BlackBerry"}, []string{}},
		},
		"Palm": {
			{417, "Palm", []int{}, []string{"Palm OS", "Palm"}, []string{}},
		},
		"Tapwave": {
			{44, "Zodiac", []int{}, []string{"Tapwave Zodiac"}, []string{}},
		},
		"Panic": {
			{381, "Playdate", []int{}, []string{"Playdate"}, []string{}},
		},
		"Tiger Telematics": {
			{474, "Gizmondo", []int{}, []string{"Gizmondo"}, []string{}},
		},
		"Legacy": {
			{55, "LegacyMobile", []int{}, []string{"Legacy Mobile Device"}, []string{}},
		},
	},
	"Arcade": {
		"Arcade": {
			{52, "Arcade", []int{}, []string{"Arcade"}, []string{}},
		},
		"Neo Geo": {
			{79, "NG_MVS", []int{}, []string{"Neo Geo", "NeoGeo"}, []string{"MVS"}},
			{80, "NG_AES", []int{}, []string{"Neo Geo", "NeoGeo"}, []string{"AES"}},
			{136, "NG_CD", []int{}, []string{"Neo Geo", "NeoGeo"}, []string{"CD"}},
			{135, "NG_64", []int{}, []string{"Hyper Neo Geo", "Hyper NeoGeo"}, []string{"64"}},
			{119, "NG_P", []int{}, []string{"Neo Geo", "NeoGeo"}, []string{"Pocket"}},
			{120, "NG_PC", []int{119}, []string{"Neo Geo", "NeoGeo"}, []string{"Pocket Color"}},
		},
		"General Instrument": {
			{140, "AY8500", []int{}, []string{"AY-3-8500"}, []string{}},
			{141, "AY8610", []int{}, []string{"AY-3-8610"}, []string{}},
			{143, "AY8760", []int{}, []string{"AY-3-8760"}, []string{}},
			{144, "AY8710", []int{}, []string{"AY-3-8710"}, []string{}},
			{145, "AY8603", []int{}, []string{"AY-3-8603"}, []string{}},
			{146, "AY8605", []int{}, []string{"AY-3-8605"}, []string{}},
			{147, "AY8606", []int{}, []string{"AY-3-8606"}, []string{}},
			{148, "AY8607", []int{}, []string{"AY-3-8607"}, []string{}},
		},
	},
	"Mini Computer": {
		"PDP": {
			{95, "PDP1", []int{}, []string{"PDP-1", "DEC PDP-1"}, []string{}},
			{103, "PDP7", []int{}, []string{"PDP-7", "DEC PDP-7"}, []string{}},
			{97, "PDP8", []int{}, []string{"PDP-8", "DEC PDP-8"}, []string{}},
			{96, "PDP10", []int{}, []string{"PDP-10", "DEC PDP-10"}, []string{}},
			{108, "PDP11", []int{}, []string{"PDP-11", "DEC PDP-11"}, []string{}},
			{98, "GT40", []int{}, []string{"DEC GT40"}, []string{}},
		},
		"HP": {
			{104, "HP2100", []int{}, []string{"HP 2100"}, []string{}},
			{105, "HP3000", []int{}, []string{"HP 3000"}, []string{}},
		},
		"CDC": {
			{109, "Cyber70", []int{}, []string{"CDC Cyber 70", "Cyber 70"}, []string{}},
		},
		"Ferranti": {
			{101, "Nimrod", []int{}, []string{"Ferranti Nimrod Computer", "Nimrod"}, []string{}},
		},
		"Cambridge": {
			{102, "EDSAC", []int{}, []string{"EDSAC"}, []string{}},
		},
		"SDS": {
			{106, "Sigma7", []int{}, []string{"SDS Sigma 7", "Sigma 7"}, []string{}},
		},
		"PLATO": {
			{110, "PLATO", []int{}, []string{"PLATO"}, []string{}},
		},
		"Imlac": {
			{111, "PDS1", []int{}, []string{"Imlac PDS-1"}, []string{}},
		},
		"Donner": {
			{85, "Model30", []int{}, []string{"Donner Model 30"}, []string{}},
		},
		"CAC": {
			{107, "CAC", []int{}, []string{"Call-A-Computer time-shared mainframe computer system", "Call-A-Computer"}, []string{}},
		},
		"Early": {
			{100, "Analogue", []int{}, []string{"Analogue electronics"}, []string{}},
			{112, "Micro", []int{}, []string{"Microcomputer"}, []string{}},
		},
	},
	"Cloud": {
		"Google": {
			{170, "Stadia", []int{}, []string{"Google Stadia", "Stadia"}, []string{}},
		},
		"OnLive": {
			{113, "OnLive", []int{}, []string{"OnLive Game System", "OnLive"}, []string{}},
		},
		"AirConsole": {
			{389, "AirConsole", []int{}, []string{"AirConsole"}, []string{}},
		},
	},
	"Educational": {
		"LeapFrog": {
			{412, "Leapster", []int{}, []string{"Leapster"}, []string{}},
			{413, "Leapster2", []int{412}, []string{"Leapster Explorer", "LeadPad Explorer", "Leapster Explorer/LeadPad Explorer"}, []string{}},
			{414, "LeapTV", []int{}, []string{"LeapTV"}, []string{}},
		},
		"VTech": {
			{439, "VSmile", []int{}, []string{"V.Smile", "VTech V.Smile"}, []string{}},
		},
		"Bandai": {
			{479, "Terebikko", []int{}, []string{"Terebikko", "See 'n Say Video Phone", "Terebikko / See 'n Say Video Phone"}, []string{}},
		},
	},
	"Homebrew": {
		"Arduboy": {
			{438, "Arduboy", []int{}, []string{"Arduboy"}, []string{}},
		},
		"Uzebox": {
			{504, "Uzebox", []int{}, []string{"Uzebox"}, []string{}},
		},
		"Elektor": {
			{505, "Elektor", []int{}, []string{"Elektor TV Games Computer"}, []string{}},
		},
		"OOParts": {
			{372, "OOParts", []int{}, []string{"OOParts"}, []string{}},
		},
	},
	"Media": {
		"Optical": {
			{238, "DVD", []int{}, []string{"DVD Player", "DVD"}, []string{}},
			{239, "BluRay", []int{238}, []string{"Blu-ray Player", "Blu-ray"}, []string{}},
		},
		"Web": {
			{82, "Web", []int{}, []string{"Web browser", "Browser"}, []string{}},
		},
	},
}
