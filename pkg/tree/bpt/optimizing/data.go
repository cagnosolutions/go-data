package optimizing

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

var nums2 = []uint64{
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

var byts = [][]byte{
	[]byte("abscess"), []byte("access"), []byte("accidents"), []byte("actors"), []byte("aegis"),
	[]byte("aerosmith"), []byte("aggregates"), []byte("agriculture"), []byte("allegation"),
	[]byte("allows"), []byte("altruism"), []byte("alvaro"), []byte("amico"),
	[]byte("animales"), []byte("antonyms"), []byte("antrim"), []byte("apiece"),
	[]byte("appleby"), []byte("applicability"), []byte("appointing"), []byte("arbiter"),
	[]byte("archaeological"), []byte("archiving"), []byte("aromatics"), []byte("arrested"),
	[]byte("artistdirect"), []byte("arxiv"), []byte("assisi"), []byte("assumptions"),
	[]byte("asustek"), []byte("athletic"), []byte("atlases"), []byte("avery"),
	[]byte("azkaban"), []byte("backstreet"), []byte("baggage"), []byte("ballads"),
	[]byte("barbed"), []byte("barenaked"), []byte("bartending"), []byte("basking"),
	[]byte("beaufort"), []byte("behaving"), []byte("bejeweled"), []byte("bestality"),
	[]byte("bestellen"), []byte("biting"), []byte("blanche"), []byte("bleeds"),
	[]byte("blindness"), []byte("bloody"), []byte("bogen"), []byte("boingo"),
	[]byte("bonding"), []byte("bookkeeping"), []byte("borer"), []byte("bottleneck"),
	[]byte("bounty"), []byte("braided"), []byte("breakdown"), []byte("breakthrough"),
	[]byte("bremer"), []byte("breton"), []byte("brill"), []byte("brushed"),
	[]byte("bufing"), []byte("bugfix"), []byte("butterflies"), []byte("byron"),
	[]byte("calais"), []byte("camping"), []byte("canonical"), []byte("capsule"),
	[]byte("carrara"), []byte("carter"), []byte("cartoon"), []byte("catwalk"),
	[]byte("cecil"), []byte("cellog"), []byte("censor"), []byte("cerruti"),
	[]byte("chalice"), []byte("champaign"), []byte("cheering"), []byte("cheerleaders"),
	[]byte("chiang"), []byte("childish"), []byte("chitty"), []byte("choice"),
	[]byte("chuck"), []byte("circulation"), []byte("clearfield"), []byte("clive"),
	[]byte("codecs"), []byte("coercive"), []byte("coincides"), []byte("commando"),
	[]byte("comming"), []byte("community"), []byte("competed"), []byte("conroy"),
	[]byte("constellations"), []byte("constitutions"), []byte("contraction"), []byte("cooling"),
	[]byte("copacabana"), []byte("copier"), []byte("copyright"), []byte("coreldraw"),
	[]byte("cornwall"), []byte("corps"), []byte("corvette"), []byte("cosmic"),
	[]byte("councilor"), []byte("coursework"), []byte("craftsmen"), []byte("crafty"),
	[]byte("creased"), []byte("cribs"), []byte("cricinfo"), []byte("critiquer"),
	[]byte("crowding"), []byte("cured"), []byte("cycle"), []byte("cygnus"),
	[]byte("cyngor"), []byte("dakota"), []byte("darkroom"), []byte("darmstadt"),
	[]byte("darwen"), []byte("darwinports"), []byte("debatable"), []byte("decks"),
	[]byte("defeats"), []byte("delineate"), []byte("delivered"), []byte("dells"),
	[]byte("denials"), []byte("denny"), []byte("dependent"), []byte("depressive"),
	[]byte("destin"), []byte("deteriorate"), []byte("developmental"), []byte("dialers"),
	[]byte("diane"), []byte("disappearance"), []byte("discipleship"), []byte("discontinue"),
	[]byte("displacements"), []byte("disseminated"), []byte("distracting"), []byte("disulfide"),
	[]byte("doorbell"), []byte("dorman"), []byte("doughnuts"), []byte("doylestown"),
	[]byte("drinks"), []byte("drowned"), []byte("dummies"), []byte("duquesne"),
	[]byte("ecuador"), []byte("edonkey"), []byte("efficient"), []byte("electors"),
	[]byte("elmer"), []byte("emoticon"), []byte("encodings"), []byte("engines"),
	[]byte("enlarging"), []byte("enterprises"), []byte("epiphone"), []byte("epistles"),
	[]byte("equality"), []byte("ericson"), []byte("ervin"), []byte("escrituras"),
	[]byte("etoile"), []byte("evidenced"), []byte("evils"), []byte("excerpted"),
	[]byte("exhaustion"), []byte("explosion"), []byte("extremities"), []byte("fable"),
	[]byte("fairbanks"), []byte("fateful"), []byte("felicia"), []byte("feline"),
	[]byte("felon"), []byte("fetching"), []byte("finishing"), []byte("flashback"),
	[]byte("foals"), []byte("fontaine"), []byte("formatted"), []byte("formerly"),
	[]byte("fostered"), []byte("fraser"), []byte("freer"), []byte("freezes"),
	[]byte("fucked"), []byte("furtherance"), []byte("gamespy"), []byte("ganglia"),
	[]byte("gases"), []byte("genocide"), []byte("geographies"), []byte("geriatrics"),
	[]byte("germans"), []byte("germantown"), []byte("germplasm"), []byte("giant"),
	[]byte("godsmack"), []byte("gordo"), []byte("gretchen"), []byte("griffith"),
	[]byte("grille"), []byte("grinding"), []byte("grisham"), []byte("growers"),
	[]byte("gryffindor"), []byte("guessing"), []byte("gunshot"), []byte("handcuffs"),
	[]byte("hangul"), []byte("health"), []byte("healthday"), []byte("heather"),
	[]byte("heine"), []byte("hemscott"), []byte("hilltop"), []byte("hinterland"),
	[]byte("holographic"), []byte("hoodwinked"), []byte("hoppers"), []byte("hotelguide"),
	[]byte("howarth"), []byte("hryvnia"), []byte("hwang"), []byte("hydropower"),
	[]byte("hygiene"), []byte("idaho"), []byte("idealism"), []byte("imaginations"),
	[]byte("imation"), []byte("impulse"), []byte("indonesian"), []byte("inherited"),
	[]byte("integrated"), []byte("interlocking"), []byte("intolerance"), []byte("intrinsically"),
	[]byte("introduction"), []byte("invaders"), []byte("investments"), []byte("invocation"),
	[]byte("irish"), []byte("irresistible"), []byte("isaiah"), []byte("isolde"),
	[]byte("italian"), []byte("iupui"), []byte("ivanhoe"), []byte("jeanne"),
	[]byte("jerusalem"), []byte("journey"), []byte("kassel"), []byte("kbyte"),
	[]byte("kennewick"), []byte("khalid"), []byte("kharagpur"), []byte("kirsty"),
	[]byte("kitchenaid"), []byte("klear"), []byte("kristin"), []byte("krystal"),
	[]byte("kuala"), []byte("kunst"), []byte("lacey"), []byte("lancer"),
	[]byte("largely"), []byte("lasting"), []byte("latitude"), []byte("leiden"),
	[]byte("lense"), []byte("lesbos"), []byte("likeness"), []byte("limassol"),
	[]byte("lineage"), []byte("linux"), []byte("lowly"), []byte("magnetometer"),
	[]byte("maier"), []byte("mailbag"), []byte("maintained"), []byte("mandated"),
	[]byte("mandel"), []byte("maneuvering"), []byte("manhattan"), []byte("mario"),
	[]byte("massage"), []byte("mavericks"), []byte("maximilian"), []byte("medina"),
	[]byte("merely"), []byte("meridia"), []byte("metastatic"), []byte("metric"),
	[]byte("microbiol"), []byte("millers"), []byte("monuments"), []byte("mornings"),
	[]byte("motherfucker"), []byte("motif"), []byte("mounds"), []byte("msnbc"),
	[]byte("mumbai"), []byte("mutex"), []byte("natural"), []byte("nearer"),
	[]byte("negotiated"), []byte("nellie"), []byte("newsletter"), []byte("newsre"),
	[]byte("ninjas"), []byte("nissan"), []byte("norbert"), []byte("norma"),
	[]byte("normandy"), []byte("northrop"), []byte("nostradamus"), []byte("novotel"),
	[]byte("numeracy"), []byte("oilers"), []byte("omnibus"), []byte("omnipotent"),
	[]byte("opaque"), []byte("ourselves"), []byte("outflow"), []byte("outlook"),
	[]byte("overkill"), []byte("oversight"), []byte("overstock"), []byte("palettes"),
	[]byte("paperwork"), []byte("parfums"), []byte("paused"), []byte("peaks"),
	[]byte("pedagogy"), []byte("periwinkle"), []byte("personalized"), []byte("persuasion"),
	[]byte("picts"), []byte("plunge"), []byte("points"), []byte("policemen"),
	[]byte("polydor"), []byte("poole"), []byte("populate"), []byte("portrayed"),
	[]byte("potash"), []byte("practically"), []byte("preached"), []byte("preamps"),
	[]byte("prestige"), []byte("prion"), []byte("probe"), []byte("processing"),
	[]byte("professorship"), []byte("progenitor"), []byte("progressively"), []byte("pronouncements"),
	[]byte("prosthetic"), []byte("pulse"), []byte("quinton"), []byte("racquet"),
	[]byte("radiated"), []byte("rahul"), []byte("rally"), []byte("reactions"),
	[]byte("reads"), []byte("rearranging"), []byte("rebates"), []byte("recording"),
	[]byte("recreated"), []byte("refinance"), []byte("rehabilitation"), []byte("rehearing"),
	[]byte("reimbursements"), []byte("reinforce"), []byte("remodeled"), []byte("reopening"),
	[]byte("replies"), []byte("reprinted"), []byte("resemblance"), []byte("resort"),
	[]byte("retires"), []byte("reviews"), []byte("revitalization"), []byte("reworked"),
	[]byte("richer"), []byte("rinse"), []byte("ripples"), []byte("roadside"),
	[]byte("rollin"), []byte("romanticism"), []byte("ruger"), []byte("santa"),
	[]byte("scaring"), []byte("scrolls"), []byte("semantically"), []byte("sername"),
	[]byte("servicemen"), []byte("sewer"), []byte("shabby"), []byte("shana"),
	[]byte("shins"), []byte("shout"), []byte("shouted"), []byte("shrunk"),
	[]byte("siberia"), []byte("signifies"), []byte("sikkim"), []byte("skins"),
	[]byte("sometimes"), []byte("sparkle"), []byte("spirit"), []byte("spouses"),
	[]byte("sprache"), []byte("stability"), []byte("stabilizers"), []byte("staircase"),
	[]byte("standing"), []byte("startseite"), []byte("stent"), []byte("stimulates"),
	[]byte("stole"), []byte("storageworks"), []byte("stove"), []byte("styrene"),
	[]byte("submarine"), []byte("substandard"), []byte("suburbs"), []byte("suicide"),
	[]byte("sunni"), []byte("surety"), []byte("surfacing"), []byte("sweethearts"),
	[]byte("synchronization"), []byte("tabled"), []byte("tackle"), []byte("tawnee"),
	[]byte("tenement"), []byte("tension"), []byte("themes"), []byte("therm"),
	[]byte("thermocouple"), []byte("thessalonians"), []byte("tiara"), []byte("tipped"),
	[]byte("tracking"), []byte("trading"), []byte("trellis"), []byte("trier"),
	[]byte("trimester"), []byte("twikivariables"), []byte("typeface"), []byte("unaids"),
	[]byte("unconditional"), []byte("unilaterally"), []byte("unmodified"), []byte("valdez"),
	[]byte("valencia"), []byte("vantec"), []byte("varadero"), []byte("variegated"),
	[]byte("versed"), []byte("vicky"), []byte("vincent"), []byte("vodka"),
	[]byte("volatility"), []byte("volkswagen"), []byte("wadsworth"), []byte("watchlist"),
	[]byte("waxing"), []byte("webservices"), []byte("weekly"), []byte("wesley"),
	[]byte("westbury"), []byte("westin"), []byte("whistling"), []byte("wildcards"),
	[]byte("winders"), []byte("windstar"), []byte("withers"), []byte("writeshield"),
	[]byte("zhang"), []byte("zipped"), []byte("zoeken"),
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

func ShuffleByts(arr *[][]byte) {
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

func PrintByts(arr [][]byte) {

	fmt.Println("Sorted:")
	for _, v := range arr {
		fmt.Printf("%q ", v)
	}
	fmt.Println()

	fmt.Println("Unsorted:")
	ShuffleByts(&arr)
	for _, v := range arr {
		fmt.Printf("%q ", v)
	}
	fmt.Println()

}
