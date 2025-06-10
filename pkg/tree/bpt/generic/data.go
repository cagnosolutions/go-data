package generic

import (
	"fmt"
	"math/rand"
	"time"
)

var nums = []uint32{
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17,
	18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33,
	34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61,
	62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89,
	90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113,
	114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 132, 133, 134, 135,
	136, 137, 138, 139, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 150, 151, 152, 153, 154, 155, 156, 157,
	158, 159, 160, 161, 162, 163, 164, 165, 166, 167, 168, 169, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179,
	180, 181, 182, 183, 184, 185, 186, 187, 188, 189, 190, 191, 192, 193, 194, 195, 196, 197, 198, 199, 200, 201,
	202, 203, 204, 205, 206, 207, 208, 209, 210, 211, 212, 213, 214, 215, 216, 217, 218, 219, 220, 221, 222, 223,
	224, 225, 226, 227, 228, 229, 230, 231, 232, 233, 234, 235, 236, 237, 238, 239, 240, 241, 242, 243, 244, 245,
	246, 247, 248, 249, 250, 251, 252, 253, 254, 255, 256, 257, 258, 259, 260, 261, 262, 263, 264, 265, 266, 267,
	268, 269, 270, 271, 272, 273, 274, 275, 276, 277, 278, 279, 280, 281, 282, 283, 284, 285, 286, 287, 288, 289,
	290, 291, 292, 293, 294, 295, 296, 297, 298, 299, 300, 301, 302, 303, 304, 305, 306, 307, 308, 309, 310, 311,
	312, 313, 314, 315, 316, 317, 318, 319, 320, 321, 322, 323, 324, 325, 326, 327, 328, 329, 330, 331, 332, 333,
	334, 335, 336, 337, 338, 339, 340, 341, 342, 343, 344, 345, 346, 347, 348, 349, 350, 351, 352, 353, 354, 355,
	356, 357, 358, 359, 360, 361, 362, 363, 364, 365, 366, 367, 368, 369, 370, 371, 372, 373, 374, 375, 376, 377,
	378, 379, 380, 381, 382, 383, 384, 385, 386, 387, 388, 389, 390, 391, 392, 393, 394, 395, 396, 397, 398, 399,
	400, 401, 402, 403, 404, 405, 406, 407, 408, 409, 410, 411, 412, 413, 414, 415, 416, 417, 418, 419, 420, 421,
	422, 423, 424, 425, 426, 427, 428, 429, 430, 431, 432, 433, 434, 435, 436, 437, 438, 439, 440, 441, 442, 443,
	444, 445, 446, 447, 448, 449, 450, 451, 452, 453, 454, 455, 456, 457, 458, 459, 460, 461, 462, 463, 464, 465,
	466, 467, 468, 469, 470, 471, 472, 473, 474, 475, 476, 477, 478, 479, 480, 481, 482, 483, 484, 485, 486, 487,
	488, 489, 490, 491, 492, 493, 494, 495, 496, 497, 498, 499, 500,
}

var strs = []string{
	"abscess", "access", "accidents", "actors", "aegis", "aerosmith", "aggregates", "agriculture", "allegation",
	"allows", "altruism", "alvaro", "amico", "animales", "antonyms", "antrim", "apiece", "appleby", "applicability",
	"appointing", "arbiter", "archaeological", "archiving", "aromatics", "arrested", "artistdirect", "arxiv",
	"assisi", "assumptions", "asustek", "athletic", "atlases", "avery", "azkaban", "backstreet", "baggage",
	"ballads", "barbed", "barenaked", "bartending", "basking", "beaufort", "behaving", "bejeweled", "bestality",
	"bestellen", "biting", "blanche", "bleeds", "blindness", "bloody", "bogen", "boingo", "bonding",
	"bookkeeping", "borer", "bottleneck", "bounty", "braided", "breakdown", "breakthrough", "bremer", "breton",
	"brill", "brushed", "bufing", "bugfix", "butterflies", "byron", "calais", "camping", "canonical", "capsule",
	"carrara", "carter", "cartoon", "catwalk", "cecil", "cellog", "censor", "cerruti", "chalice", "champaign",
	"cheering", "cheerleaders", "chiang", "childish", "chitty", "choice", "chuck", "circulation", "clearfield",
	"clive", "codecs", "coercive", "coincides", "commando", "comming", "community", "competed", "conroy",
	"constellations", "constitutions", "contraction", "cooling", "copacabana", "copier", "copyright", "coreldraw",
	"cornwall", "corps", "corvette", "cosmic", "councilor", "coursework", "craftsmen", "crafty", "creased",
	"cribs", "cricinfo", "critiquer", "crowding", "cured", "cycle", "cygnus", "cyngor", "dakota", "darkroom",
	"darmstadt", "darwen", "darwinports", "debatable", "decks", "defeats", "delineate", "delivered", "dells",
	"denials", "denny", "dependent", "depressive", "destin", "deteriorate", "developmental", "dialers", "diane",
	"disappearance", "discipleship", "discontinue", "displacements", "disseminated", "distracting", "disulfide",
	"doorbell", "dorman", "doughnuts", "doylestown", "drinks", "drowned", "dummies", "duquesne", "ecuador",
	"edonkey", "efficient", "electors", "elmer", "emoticon", "encodings", "engines", "enlarging", "enterprises",
	"epiphone", "epistles", "equality", "ericson", "ervin", "escrituras", "etoile", "evidenced", "evils",
	"excerpted", "exhaustion", "explosion", "extremities", "fable", "fairbanks", "fateful", "felicia", "feline",
	"felon", "fetching", "finishing", "flashback", "foals", "fontaine", "formatted", "formerly", "fostered",
	"fraser", "freer", "freezes", "fucked", "furtherance", "gamespy", "ganglia", "gases", "genocide", "geographies",
	"geriatrics", "germans", "germantown", "germplasm", "giant", "godsmack", "gordo", "gretchen", "griffith",
	"grille", "grinding", "grisham", "growers", "gryffindor", "guessing", "gunshot", "handcuffs", "hangul",
	"health", "healthday", "heather", "heine", "hemscott", "hilltop", "hinterland", "holographic", "hoodwinked",
	"hoppers", "hotelguide", "howarth", "hryvnia", "hwang", "hydropower", "hygiene", "idaho", "idealism",
	"imaginations", "imation", "impulse", "indonesian", "inherited", "integrated", "interlocking", "intolerance",
	"intrinsically", "introduction", "invaders", "investments", "invocation", "irish", "irresistible", "isaiah",
	"isolde", "italian", "iupui", "ivanhoe", "jeanne", "jerusalem", "journey", "kassel", "kbyte", "kennewick",
	"khalid", "kharagpur", "kirsty", "kitchenaid", "klear", "kristin", "krystal", "kuala", "kunst", "lacey",
	"lancer", "largely", "lasting", "latitude", "leiden", "lense", "lesbos", "likeness", "limassol", "lineage",
	"linux", "lowly", "magnetometer", "maier", "mailbag", "maintained", "mandated", "mandel", "maneuvering",
	"manhattan", "mario", "massage", "mavericks", "maximilian", "medina", "merely", "meridia", "metastatic",
	"metric", "microbiol", "millers", "monuments", "mornings", "motherfucker", "motif", "mounds", "msnbc",
	"mumbai", "mutex", "natural", "nearer", "negotiated", "nellie", "newsletter", "newsre", "ninjas", "nissan",
	"norbert", "norma", "normandy", "northrop", "nostradamus", "novotel", "numeracy", "oilers", "omnibus",
	"omnipotent", "opaque", "ourselves", "outflow", "outlook", "overkill", "oversight", "overstock", "palettes",
	"paperwork", "parfums", "paused", "peaks", "pedagogy", "periwinkle", "personalized", "persuasion", "picts",
	"plunge", "points", "policemen", "polydor", "poole", "populate", "portrayed", "potash", "practically",
	"preached", "preamps", "prestige", "prion", "probe", "processing", "professorship", "progenitor",
	"progressively", "pronouncements", "prosthetic", "pulse", "quinton", "racquet", "radiated", "rahul",
	"rally", "reactions", "reads", "rearranging", "rebates", "recording", "recreated", "refinance",
	"rehabilitation", "rehearing", "reimbursements", "reinforce", "remodeled", "reopening", "replies",
	"reprinted", "resemblance", "resort", "retires", "reviews", "revitalization", "reworked", "richer",
	"rinse", "ripples", "roadside", "rollin", "romanticism", "ruger", "santa", "scaring", "scrolls",
	"semantically", "sername", "servicemen", "sewer", "shabby", "shana", "shins", "shout", "shouted",
	"shrunk", "siberia", "signifies", "sikkim", "skins", "sometimes", "sparkle", "spirit", "spouses",
	"sprache", "stability", "stabilizers", "staircase", "standing", "startseite", "stent", "stimulates",
	"stole", "storageworks", "stove", "styrene", "submarine", "substandard", "suburbs", "suicide", "sunni",
	"surety", "surfacing", "sweethearts", "synchronization", "tabled", "tackle", "tawnee", "tenement",
	"tension", "themes", "therm", "thermocouple", "thessalonians", "tiara", "tipped", "tracking", "trading",
	"trellis", "trier", "trimester", "twikivariables", "typeface", "unaids", "unconditional", "unilaterally",
	"unmodified", "valdez", "valencia", "vantec", "varadero", "variegated", "versed", "vicky", "vincent",
	"vodka", "volatility", "volkswagen", "wadsworth", "watchlist", "waxing", "webservices", "weekly", "wesley",
	"westbury", "westin", "whistling", "wildcards", "winders", "windstar", "withers", "writeshield", "zhang",
	"zipped", "zoeken",
}

func ShuffleInts(arr *[]uint32) {
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator
	n := len(*arr)
	for i := n - 1; i > 0; i-- {
		j := rand.Intn(i + 1)                       // Get a random index from 0 to i
		(*arr)[i], (*arr)[j] = (*arr)[j], (*arr)[i] // Swap elements
	}
}

func ShuffleStrs(arr *[]string) {
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator
	n := len(*arr)
	for i := n - 1; i > 0; i-- {
		j := rand.Intn(i + 1)                       // Get a random index from 0 to i
		(*arr)[i], (*arr)[j] = (*arr)[j], (*arr)[i] // Swap elements
	}
}

func PrintInts(arr []uint32) {

	fmt.Println("Sorted:")
	for _, v := range arr {
		fmt.Printf("%d ", v)
	}
	fmt.Println()

	fmt.Println("Unsorted:")
	ShuffleInts(&arr)
	for _, v := range arr {
		fmt.Printf("%d ", v)
	}
	fmt.Println()

}

func PrintStrs(arr []string) {

	fmt.Println("Sorted:")
	for _, v := range arr {
		fmt.Printf("%q ", v)
	}
	fmt.Println()

	fmt.Println("Unsorted:")
	ShuffleStrs(&arr)
	for _, v := range arr {
		fmt.Printf("%q ", v)
	}
	fmt.Println()

}
