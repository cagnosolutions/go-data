package main

import (
	"bufio"
	"bytes"
	"compress/flate"
	"compress/lzw"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/cagnosolutions/go-data/pkg/util"
)

func invertCase(ch byte) byte {
	// ch ^= 0x20 // the number 32
	return ch ^ 0x20
}

func isLower(ch byte) bool {
	return (ch | 0x20) == 0
}

func isUpper(ch byte) bool {
	return (ch & 0x20) == 0
}

func toLower(ch *byte) {
	*ch |= 0x20
}

func toUpper(ch *byte) {
	*ch &= 0x20
}

func main() {
	dm := ReadAndCompress(firstNames, lzwComp)
	for k, v := range dm {
		fmt.Printf("%s\t--->\t%s\n", k, v)
	}
}

var lzwComp = func(data []byte) []byte {
	var buf bytes.Buffer
	w := lzw.NewWriter(&buf, lzw.LSB, 8)
	_, err := w.Write(data)
	if err != nil {
		panic(err)
	}
	err = w.Close()
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

var flateComp = func(data []byte) []byte {
	var buf bytes.Buffer
	w, err := flate.NewWriter(&buf, flate.BestCompression)
	if err != nil {
		panic(err)
	}
	_, err = w.Write(data)
	if err != nil {
		panic(err)
	}
	err = w.Flush()
	if err != nil {
		panic(err)
	}
	err = w.Close()
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func ReadAndCompress(file string, fn compress) map[string]string {
	r, err := os.Open(filepath.Join(basePath, file))
	if err != nil {
		panic(err)
	}
	defer func(r *os.File) {
		err := r.Close()
		if err != nil {
			panic(err)
		}
	}(r)
	dm := make(map[string]string)
	br := bufio.NewReader(r)
	for {
		line, err := br.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		if bytes.Equal(line, []byte{'\n'}) {
			break
		}
		line = line[:len(line)-1]
		// add the original text as the key, and the compressed text as the value
		orig := string(line)
		comp := string(fn(line))
		dm[orig] = comp
	}
	return dm
}

type compress func(data []byte) []byte

func runNgramChecker() {
	// check for any grams in the gettysburg address
	data := []byte(util.GettysburgAddress)

	// see how many n-grams we can find
	ngs := make(map[string]int)

	// make our ngram function
	fn := func(g []byte, beg, end int) bool {
		// if _, match := bigram[string(g)]; match {
		ngs[string(g)]++
		// }
		return true
	}

	// run our ngram scanner
	err := NgramScanner(2, data, fn)
	if err != nil {
		panic(err)
	}

	// see what matches we have (occurred, occurrence)
	for gram, occurrence := range ngs {
		fmt.Printf("Ngram %q occurred %d times\n", gram, occurrence)
	}
}

type NgramFn func(g []byte, beg, end int) bool

// NgramScanner will iterate a string and emit n-grams into the provided function
// NGramFn. NGram takes n, which is the gram size you wish to isolate. If len(p)
// (which is the data provided) is less than the n size provided, an error will
// be returned.
func NgramScanner(n int, p []byte, fn NgramFn) error {
	if len(p) < n {
		return errors.New("the provided data is not large enough")
	}
	if fn == nil {
		return errors.New("the provided NgramFn must not be nil")
	}
	for i, j := 0, n; j < len(p); i, j = i+1, j+1 {
		for k := i; k < j; k++ {
			if isUpper(p[k]) {
				toLower(&p[k])
			}
		}
		if !fn(p[i:j], i, j) {
			break
		}
	}
	return nil
}

var gram = map[string]byte{
	"!":  0x00,
	"@":  0x00,
	"#":  0x00,
	"$":  0x00,
	"%":  0x00,
	"^":  0x00,
	"&":  0x00,
	"*":  0x00,
	"(":  0x00,
	")":  0x00,
	" ":  0x00,
	"_":  0x00,
	"+":  0x00,
	"-":  0x00,
	"=":  0x00,
	"\\": 0x00,
	"|":  0x00,
	"`":  0x00,
	"~":  0x00,
	"/":  0x00,
	"?":  0x00,
	".":  0x00,
	",":  0x00,
	"\"": 0x00,
	"'":  0x00,
	";":  0x00,
	":":  0x00,
	"<":  0x00,
	">":  0x00,
	"{":  0x00,
	"}":  0x00,
	"[":  0x00,
	"]":  0x00,
}

var bigram = map[string]byte{
	"th": 0x00,
	"he": 0x01,
	"in": 0x02,
	"er": 0x03,
	"an": 0x04,
	"re": 0x05,
	"nd": 0x06,
	"on": 0x07,
	"en": 0x08,
	"at": 0x09,
	"ou": 0x0a,
	"ed": 0x0b,
	"ha": 0x0c,
	"to": 0x0d,
	"or": 0x0e,
	"it": 0x0f,
	"is": 0x10,
	"hi": 0x11,
	"es": 0x12,
	"ng": 0x13,
	"nt": 0x14,
	"ti": 0x15,
	"se": 0x16,
	"ar": 0x17,
	"al": 0x18,
	"te": 0x19,
	"co": 0x1a,
	"de": 0x1b,
	"ra": 0x1c,
	"et": 0x1d,
	"sa": 0x1e,
	"em": 0x1f,
	"ro": 0x20,
}

var trigram = map[string]byte{
	"the": 0x21,
	"and": 0x22,
	"ing": 0x23,
	"her": 0x24,
	"hat": 0x25,
	"his": 0x26,
	"tha": 0x27,
	"ere": 0x28,
	"for": 0x29,
	"ent": 0x2a,
	"ion": 0x2b,
	"ter": 0x2c,
	"was": 0x2d,
	"you": 0x2e,
	"ith": 0x2f,
	"ver": 0x30,
	"all": 0x31,
	"wit": 0x32,
	"thi": 0x33,
	"tio": 0x34,
	"nde": 0x35,
	"has": 0x36,
	"nce": 0x37,
	"edt": 0x38,
	"tis": 0x39,
	"oft": 0x3a,
	"sth": 0x3b,
	"men": 0x3c,
}

var quadgram = map[string]byte{
	"that": 0x00,
	"ther": 0x00,
	"with": 0x00,
	"tion": 0x00,
	"here": 0x00,
	"ould": 0x00,
	"ight": 0x00,
	"have": 0x00,
	"hich": 0x00,
	"whic": 0x00,
	"this": 0x00,
	"thin": 0x00,
	"they": 0x00,
	"atio": 0x00,
	"ever": 0x00,
	"from": 0x00,
	"ough": 0x00,
	"were": 0x00,
	"hing": 0x00,
	"ment": 0x00,
}

var (
	basePath   = "cmd/key_compression"
	firstNames = "sorted-first-names.txt"
	lastNames  = "sorted-last-names.txt"
	emails     = "sorted-emails.txt"
	domains    = "sorted-domains.txt"
	usernames  = "sorted-usernames.txt"
	companies  = "sorted-companies.txt"
)

func compressAndSortFile(filename string) []string {
	r, err := os.Open(filepath.Join(basePath, filename))
	if err != nil {
		panic(err)
	}
	defer func(r *os.File) {
		err := r.Close()
		if err != nil {
			panic(err)
		}
	}(r)

	var ori []string
	var set []string
	br := bufio.NewReader(r)
	for {
		line, err := br.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		if bytes.Equal(line, []byte{'\n'}) {
			break
		}
		line = line[:len(line)-1]
		ori = append(ori, string(line))
		set = append(set, string(simpleCompress(line, 4)))
	}
	// sort.Strings(set)
	for i, s := range set {
		fmt.Printf("%q -> %q\n", ori[i], s)
	}

	// return sorted set
	return set
}

func simpleCompress(b []byte, max int) []byte {
	if len(b) <= max {
		return b
	}
	dat := make([]byte, max, max)
	fix := max / 2
	copy(dat[:fix], b[:fix])
	copy(dat[fix:], b[len(b)-fix:])
	return dat
}

func generate() {

	// open file for writing
	fd, err := os.OpenFile("cmd/key_compression/sorted-data.txt", os.O_RDWR|os.O_SYNC, 0644)
	if err != nil {
		panic(err)
	}

	// range data map
	for kind, list := range data {
		// split and sort data from this list
		fmt.Printf("Splitting and sorting %q data...\n", kind)

		// remove any duplicates, and add to new array to be sorted
		unique := map[string]struct{}{}
		var count int
		for _, line := range strings.Split(list, "\n") {
			unique[line] = struct{}{}
			count++
		}
		fmt.Printf("Removed %d duplicate entries\n", count-len(unique))

		// add the unique entries to a new set
		set := make([]string, len(unique))
		for line := range unique {
			fmt.Println(len(line), line)
			set = append(set, line)
		}

		// sort the set
		sort.Strings(set)

		// range the sorted data and write to file
		for _, line := range set {
			_, err = fd.WriteString(line + "\n")
			if err != nil {
				panic(err)
			}
		}
		// flush write to disk
		err = fd.Sync()
		if err != nil {
			panic(err)
		}
		time.Sleep(3 * time.Second)
	}

	// close the file
	err = fd.Close()
	if err != nil {
		panic(err)
	}
}

var data = map[string]string{
	"company": `
Roomm
Voonyx
Cogibox
Tagtune
Tanoodle
Bluezoom
Yata
Demizz
Shufflester
Centizu
Thoughtstorm
Tagchat
Mydeo
Muxo
Skiba
Edgetag
Feedfish
Jaxworks
Blognation
Yoveo
Avavee
Gabcube
Tagcat
Trilith
Tagfeed
Flashspan
Wikivu
Shuffledrive
Ainyx
Demivee
Twinder
Yacero
Centimia
Skipfire
Bluezoom
Zooveo
Dynazzy
Yodo
Thoughtworks
Eazzy
Photospace
Fadeo
Zoombox
Mybuzz
Demimbu
Oodoo
Devpoint
Jatri
Devcast
Riffwire
Yacero
Edgewire
Twitterlist
Roombo
Ntags
Wikizz
DabZ
Gigazoom
Leexo
Devbug
Topiclounge
Wikido
Yadel
Divape
Kazu
Voonix
Centidel
Shufflebeat
Skiptube
Wikido
Viva
Feedbug
Zooxo
Kaymbo
Realbuzz
Yombu
Vipe
Meejo
Twimbo
Browsedrive
Eamia
Fiveclub
Realblab
Babbleset
Meembee
Skinte
Meeveo
Devcast
Fivebridge
Myworks
Skyvu
Plambee
Feedfish
Fanoodle
Realblab
Wikido
Fivespan
Eire
Yoveo
Shuffledrive
Zooxo
Livefish
Wikizz
Voonyx
Dynazzy
Livepath
Tanoodle
Roombo
Brightdog
Oozz
Feedfish
Buzzshare
Photojam
Flashspan
Youbridge
Twitterbridge
Youspan
Gabvine
Riffpedia
Feedmix
Jaxworks
Oloo
Cogilith
Talane
Gabtune
Jayo
Jetwire
Camimbo
Wikizz
Voonte
Brightbean
Vimbo
Plambee
Photolist
Oozz
Tavu
Kwideo
Mudo
Twitterworks
Linklinks
Jayo
Zazio
Wikizz
Fadeo
BlogXS
Photobug
Wikizz
Dabjam
Wikizz
Flipstorm
Riffwire
Kaymbo
Mudo
Flipopia
Aivee
Skiptube
Ainyx
Avamm
Blogtags
Dynazzy
Skilith
Twinder
Meetz
Quatz
Rhynyx
Eayo
Meevee
Realcube
Wikido
Plajo
Gevee
Trudoo
Kare
Tekfly
Vipe
Yambee
Layo
Minyx
Dabshots
Yakijo
Rhyloo
Gigashots
Yodoo
Gabspot
Cogilith
Thoughtstorm
Livetube
Jaxspan
Quire
Thoughtsphere
Avavee
Ntags
Skippad
Topdrive
Camido
Shufflebeat
Twinder
Trudoo
Podcat
Zoonoodle
Jaxnation
Wikivu
Demimbu
Quatz
Buzzdog
Rhybox
Wordpedia
Eidel
Youopia
Lazz
Skinder
Zoombox
Skimia
Bubblebox
Yodoo
Topicware
Ooba
Lajo
Pixoboo
Plambee
Browsebug
Skimia
Cogilith
Mymm
Avamba
Photobean
Vitz
Zoonoodle
Feedfish
Eadel
Camido
Yombu
Meembee
Meembee
Skynoodle
Skinder
Skipstorm
Flipopia
Teklist
Meemm
Fatz
Lajo
Twimm
Shufflebeat
Demimbu
Oyoyo
Youspan
Jayo
Skajo
Fivespan
Shufflester
Bluejam
Kare
Mybuzz
Midel
Thoughtblab
Brainverse
Trunyx
Centizu
Cogilith
Feedfire
Quinu
Cogidoo
Cogibox
Browsecat
Aimbu
Eabox
Gigazoom
Layo
Jatri
Trupe
Mydeo
Thoughtbeat
Muxo
Chatterpoint
Youspan
Topicstorm
JumpXS
Babbleblab
Twiyo
Zava
LiveZ
Aimbo
Gabtune
Roodel
Dabshots
Chatterbridge
Fiveclub
Zooveo
Pixoboo
Linkbuzz
Jabbersphere
Wikizz
Shufflester
Midel
Youfeed
Skinder
Zoonoodle
Pixonyx
Demimbu
`,

	"domain": `
yelp.com
de.vu
kickstarter.com
google.pl
myspace.com
sitemeter.com
miitbeian.gov.cn
latimes.com
devhub.com
yahoo.co.jp
springer.com
jigsy.com
walmart.com
globo.com
ucoz.ru
surveymonkey.com
bizjournals.com
tinyurl.com
is.gd
dmoz.org
epa.gov
youtu.be
icq.com
loc.gov
cdc.gov
feedburner.com
joomla.org
mlb.com
i2i.jp
jigsy.com
digg.com
squidoo.com
umn.edu
msu.edu
answers.com
intel.com
themeforest.net
last.fm
slate.com
t.co
about.me
whitehouse.gov
bloglines.com
de.vu
census.gov
google.com.hk
wikimedia.org
surveymonkey.com
feedburner.com
nbcnews.com
ovh.net
qq.com
pinterest.com
ezinearticles.com
tinyurl.com
lulu.com
myspace.com
wordpress.com
desdev.cn
facebook.com
slate.com
bing.com
auda.org.au
zdnet.com
pinterest.com
fda.gov
bizjournals.com
mlb.com
chicagotribune.com
cloudflare.com
walmart.com
phpbb.com
xrea.com
chicagotribune.com
booking.com
t.co
miibeian.gov.cn
amazonaws.com
psu.edu
altervista.org
telegraph.co.uk
salon.com
adobe.com
scribd.com
t.co
mashable.com
ehow.com
usgs.gov
hp.com
house.gov
dot.gov
bing.com
ftc.gov
163.com
google.cn
arstechnica.com
ca.gov
paginegialle.it
washingtonpost.com
lycos.com
cdbaby.com
oaic.gov.au
fema.gov
paypal.com
intel.com
vk.com
techcrunch.com
nymag.com
reference.com
sciencedirect.com
paypal.com
taobao.com
squarespace.com
apache.org
walmart.com
godaddy.com
elpais.com
seattletimes.com
printfriendly.com
1688.com
godaddy.com
patch.com
phpbb.com
unc.edu
example.com
mtv.com
deviantart.com
google.it
ebay.com
rambler.ru
earthlink.net
mit.edu
blogger.com
telegraph.co.uk
wikimedia.org
slideshare.net
de.vu
hao123.com
sakura.ne.jp
opera.com
spiegel.de
msn.com
desdev.cn
uol.com.br
google.com.hk
de.vu
storify.com
weebly.com
washingtonpost.com
opera.com
sourceforge.net
dot.gov
odnoklassniki.ru
etsy.com
loc.gov
mozilla.org
google.cn
who.int
discovery.com
berkeley.edu
123-reg.co.uk
comcast.net
patch.com
stumbleupon.com
discovery.com
yolasite.com
state.gov
google.nl
moonfruit.com
google.ca
hp.com
engadget.com
discuz.net
jigsy.com
phpbb.com
dyndns.org
ibm.com
jiathis.com
unc.edu
simplemachines.org
godaddy.com
squidoo.com
vk.com
hhs.gov
blinklist.com
ihg.com
techcrunch.com
discovery.com
instagram.com
hp.com
biblegateway.com
usnews.com
google.ru
ustream.tv
usatoday.com
toplist.cz
amazon.co.jp
columbia.edu
google.ca
ifeng.com
hc360.com
ucoz.ru
vinaora.com
clickbank.net
geocities.jp
youtube.com
nydailynews.com
biglobe.ne.jp
exblog.jp
dyndns.org
nsw.gov.au
umich.edu
yolasite.com
toplist.cz
ustream.tv
vimeo.com
epa.gov
cnn.com
kickstarter.com
hhs.gov
phoca.cz
feedburner.com
vkontakte.ru
furl.net
i2i.jp
hud.gov
photobucket.com
pen.io
noaa.gov
msn.com
amazon.com
chron.com
liveinternet.ru
gizmodo.com
wufoo.com
ucsd.edu
dion.ne.jp
huffingtonpost.com
networkadvertising.org
t-online.de
cdc.gov
ft.com
chicagotribune.com
goodreads.com
smh.com.au
salon.com
github.com
aboutads.info
discuz.net
va.gov
blogtalkradio.com
blogs.com
unc.edu
gizmodo.com
odnoklassniki.ru
eepurl.com
delicious.com
alexa.com
theglobeandmail.com
sohu.com
sohu.com
fastcompany.com
tuttocitta.it
sohu.com
etsy.com
ucoz.com
ucla.edu
vistaprint.com
berkeley.edu
sogou.com
timesonline.co.uk
businesswire.com
toplist.cz
cdbaby.com
github.io
samsung.com
toplist.cz
pbs.org
ucsd.edu
biglobe.ne.jp
pbs.org
va.gov
weather.com
abc.net.au
hao123.com
wunderground.com
shutterfly.com
cdbaby.com
flavors.me
nps.gov
mit.edu
wufoo.com
etsy.com
loc.gov
businessinsider.com
miibeian.gov.cn
jalbum.net
bbc.co.uk
istockphoto.com
ox.ac.uk
`,

	"username": `
randrivot0
fgifford1
cmaud2
veggleston3
fbreache4
agatsby5
dparkinson6
awraith7
bmacguire8
pbenes9
rbirtleya
cyeatsb
elombardc
pstollenbeckd
utrimmee
vfarnhillf
amutlowg
bescreeth
ceasoni
tpantecostj
lminchik
mgrishelyovl
gbromagem
obaudinetn
kgillmoro
cdowthwaitep
bmellemq
nhurlr
valexandres
pdeernesst
doakhillu
ashillv
swestwellw
rmiddleweekx
aalasdairy
ecortesz
gbouchier10
cradcliffe11
fmckomb12
mspringett13
kcary14
cmactrustam15
dmeeron16
sgriswood17
crhoddie18
anono19
fduckitt1a
aortler1b
grider1c
epluthero1d
tmcavey1e
dcypler1f
flympany1g
ckenningley1h
pcrebo1i
mhartopp1j
tcasebourne1k
ktout1l
sjess1m
mdabourne1n
sradage1o
ngronou1p
lbiggerdike1q
mdonson1r
zheathorn1s
ahonack1t
tcadding1u
rdumbreck1v
trasor1w
lhoulston1x
jbarker1y
shadgraft1z
kisack20
ckyberd21
mblacksell22
sscallon23
blittleover24
rjordine25
sspillane26
ilaterza27
lpoacher28
fonians29
krymer2a
kbreens2b
jandrejevic2c
ipinel2d
afraser2e
mstarie2f
epitts2g
meck2h
gokenden2i
jblurton2j
dfilgate2k
bhuortic2l
mpryor2m
mscutcheon2n
aburdikin2o
cdorricott2p
hpeller2q
rlevine2r
ahalfacree0
ocurzey1
jchalker2
tatto3
kcolling4
abrookesbie5
mbremley6
hduckworth7
mlambell8
millesley9
dmatzena
wcharlwoodb
ifetteplacec
kcolleerd
sblaxtere
nverbruggenf
dbeveridgeg
pbarffordh
gjoseferi
hpratleyj
kroderk
vgormanl
agrasm
besomen
bsebrooko
gcosinp
ctyghtq
hpimr
caynsleys
dkirrenst
lsaenzu
cprinnv
jcoolw
vtimlettx
eslaymakery
lbalshenz
gsuttell10
mswale11
daleso12
wviggers13
rpaunton14
bmeade15
waaronsohn16
eellingsworth17
nprandini18
mmathias19
pitzkowicz1a
stoffler1b
itilburn1c
lchaman1d
ckall1e
ccliburn1f
hmuddle1g
cgarley1h
aupham1i
noattes1j
bsheere1k
dwastall1l
mduro1m
ggarford1n
lsailes1o
jcahn1p
ldukelow1q
sbarnwall1r
kjohanssen1s
hphilippon1t
bparkin1u
ddosdale1v
edagon1w
bhurley1x
afullerton1y
rlabbett1z
hcharte20
redworthy21
mkennford22
esyfax23
ahelis24
ycustard25
preignard26
bskones27
szarfai28
cantoniat29
vgavagan2a
mcasterot2b
hgrenshiels2c
ekaye2d
tmossman2e
lgirogetti2f
doldknowe2g
esigfrid2h
lmulvihill2i
sdunkerley2j
srowe2k
mnorvel2l
kstienham2m
llowey2n
rbinestead2o
bmaro2p
mleverich2q
lmordie2r
hpyford0
ebending1
tmcclymond2
ctiuit3
jlehrahan4
corteu5
lsebrens6
mlakeman7
tpepall8
rcurnokk9
tbernardta
dgomerb
hfihellyc
sthickensd
aruddimane
fatteridgef
gsparshettg
glowingsh
cgaffeyi
awhitesonj
fcoddk
dlasslettl
jgreenhowm
cprivern
dackermanno
hambrosinip
adaniauq
akynsonr
amarshallsays
abramsomt
jsavaageu
cmilbournv
glockhurstw
dstarkeyx
akeersy
lkitchenerz
gduny10
kcastelyn11
lhuot12
tleggitt13
zmagister14
mrozec15
jhowling16
cjacmar17
ppatshull18
ahackelton19
mhumpatch1a
hmills1b
cpiper1c
wbruinemann1d
rphysic1e
mbenedicto1f
jpacitti1g
vpeppard1h
lmuffen1i
schilderhouse1j
sterne1k
blascell1l
kgurden1m
svalsler1n
hsweetmore1o
bmackellar1p
tbritee1q
dblencowe1r
ldartan1s
jdarlaston1t
csilber1u
kcampanelli1v
sbrundrett1w
ehinchcliffe1x
mfiles1y
whastwell1z
leagle20
ksouthern21
mwestoff22
scolten23
gthickett24
cslainey25
dpawlata26
elangmead27
lshyre28
ephibb29
otalmadge2a
rcockshut2b
ebrighouse2c
egulleford2d
mrowsel2e
amcteggart2f
bratnage2g
radam2h
fkaysor2i
bhaggard2j
smichel2k
eraff2l
vdunham2m
tbrookhouse2n
egibberd2o
bsouza2p
mtaggerty2q
ebasson2r
`,

	"lastName": `
Jozsef
Yeowell
Trenbay
Barden
Bavidge
Ansett
Shellshear
Kerrich
Davies
Fearenside
de Bullion
Zanicchi
Clancy
Krysztowczyk
Scouller
Norewood
Casbolt
Londesborough
Daniely
Botwright
Pattington
Pree
Bartell
Arnecke
Austen
Shoebotham
Caudell
Crosi
Berzin
Marriot
Churms
Fackrell
Kleinberer
Wittman
Tiffney
Roser
Rigate
Davidovsky
Scollan
Plumridge
Fishbourn
Langthorn
Aldie
Bortolomei
McAvey
Cordery
Pala
Comber
Larwell
Trowsdall
Wibrew
Skentelbery
Order
Kembrey
Shine
Wehner
Griffitts
Gardner
Pittendreigh
Habbin
Fleckney
Spurritt
Sutherns
Ferrierio
Backshell
Blount
Barz
Chill
Hamlett
Weinberg
Ciciari
Hammand
Hengoed
Norledge
Blowfield
Gostage
Heskey
Scedall
Trethewey
Jauncey
McCree
Andreutti
Labro
Caveau
Salla
Bartlet
Fitzhenry
Phippard
Otteridge
Kemm
Follitt
Crawford
O'Curran
Shrubsall
Minogue
Slogrove
Polfer
Axelbee
Moehle
Campagne
Flacke
Schooling
Cheese
Robottom
Casiero
Milius
Wellen
Battman
Gershom
Harbar
Pierpoint
Newbery
Fruchon
Juza
Grealy
Pabel
Doyle
Ashdown
Matiebe
Chaman
Kahler
Zapata
Loadman
Sirman
Dulling
Tanti
Comberbach
McKinney
McGilleghole
Grey
Eliaz
Bourchier
Cotty
Fogel
Harber
Peer
Tween
Curryer
Bousfield
Dunkinson
Blagburn
Dunning
Grimsdyke
Eastup
Daynter
Harefoot
Ventum
Godbert
Josskowitz
Meach
Dowse
Pourveer
Gasnoll
Hallahan
Roache
Martelet
Godbolt
Petrelli
O'Hartagan
Tzuker
Rodell
Fairman
Mullen
Tallquist
McCuffie
Sidery
Brittles
Enrique
Martin
O'Neary
Shall
Spiniello
Madle
Soughton
Dursley
Blakely
Matusov
Sissot
Veazey
Hoffmann
Ziemsen
Ulster
Menco
Dungey
Toth
Reddle
MacDonough
Kennewell
MacGaughie
Tolchard
Pendrid
Rozet
Wherton
Rawlcliffe
Allred
Mounter
Knott
Varley
Noke
Kauscher
Hurn
Daughton
Thompson
Guidone
Rummery
Scrooby
Domleo
Dobbie
Hilliam
Rigard
Reilinger
Swadlen
Tomney
Hirschmann
Clementel
Skoate
Garlic
Baum
Kadd
McClintock
Oxtiby
Broadley
Menham
Ianizzi
Leere
Fellowes
Franklyn
Tott
Niesing
Maskill
Conrart
McKibbin
McCuthais
Byart
Rubbens
Kitchenside
Livingston
Crosse
Clemow
Denny
Butterworth
Lewis
Dukelow
Dunstan
Ghidelli
Jirek
Yu
Trudgeon
Blaxill
Harnetty
Jacqueminot
Maudett
Effemy
Proffitt
Santore
Jessett
Yurasov
Bercevelo
Sellman
Simanek
Dotson
Gipson
Pylkynyton
Callam
De Cruz
Smeuin
Hatrey
Grealy
Corhard
Hexter
Origin
Bredes
Faye
Emer
Newark
Ede
Vigurs
Duigan
Feldmesser
Matuszak
Gerrie
Wathall
Kennard
Aleksidze
Gipps
Josipovic
Frosch
Loughan
Fettes
Dyke
Driutti
Harroway
Core
Mellish
Hyman
Malenfant
Wintringham
Dohmann
Chezier
Realy
`,

	"firstName": `
Suzanne
Vere
Torey
Shannon
Ring
Goran
Loralie
Alejandro
Bartholomeo
Reid
Gilles
Merrilee
Theressa
Bunni
Justen
Gifford
Fredek
Amity
Lorilyn
Avigdor
Wilhelmina
Fanya
Donelle
Lorri
Seth
Tristam
Elwin
Silvanus
Pierce
Meghann
Roland
Karine
Heidi
Calypso
Constantina
Angelo
Muffin
Laurel
Farrell
Korey
Lurlene
Ricardo
Trudy
Carlynn
Lane
Miof mela
Grange
Kerrill
Ward
Darrelle
Glyn
Klemens
Axe
Lorain
Brice
Sheba
Jolynn
Andria
Edgar
Leandra
Wendall
Hulda
Addie
Shurwood
Ricard
Eugine
Pat
Evangelin
Alana
Ingra
Townsend
Roxanna
Nari
Dredi
Louisa
Ginger
Marlene
Savina
Willabella
Maxie
Amii
Dania
Lissie
Ruddie
Jenna
Morlee
Spenser
Abbe
Sheelah
Northrup
Thaddus
Archie
Lynne
Gabriel
Adelbert
Lynnette
Maisey
Hilario
Moises
Maudie
Rosene
Cornie
Curtis
Maryann
Rockwell
Martyn
Briney
Phillis
Bartlet
Effie
Karena
Cynthia
Augusto
Thaxter
Britta
Kari
Jewel
Raquela
Parry
Mattheus
Ahmed
Batholomew
Edd
Lorene
Rodger
Jobi
Mychal
Arabele
Isa
Oby
Kalila
Lynett
Wynn
Anne-marie
Dilan
Maximilien
Wadsworth
Lemmie
Deloris
Ethelda
Geri
Gregoire
Gustav
Lorrie
Alvera
Rosanne
Arnie
Anallese
Sara-ann
Dayle
Carmon
Randy
Genvieve
Anderson
Lonnie
Gayle
Harlan
Garold
Raeann
Reagen
Belinda
Thayne
Kristi
Leonora
Hyacinthia
Catlin
Kirby
Jerrome
Maynard
Jessalyn
Hamlen
Denny
Anni
Britni
Ailee
Maiga
Deeann
Aurea
Reuben
Rickie
Laverne
Morganne
Merle
Yettie
Nanine
Kally
Hans
Minetta
Enrique
Eleen
Harold
Owen
Orran
Kayla
Errol
Crissy
Carmencita
Gaynor
Happy
Deny
Roseline
Philippine
Doralyn
Lyn
Stephanie
Reese
Florenza
Dav
Tomkin
Dyana
Marney
Maggi
Sidonia
Janet
Dorry
Vernen
Kathi
Mathe
Gav
Becka
Linnea
Sybil
Sydelle
Bondie
Norman
Swen
Janine
Helene
Merridie
Neda
Eileen
Gwynne
Stacia
Linda
Bessy
Petra
Riki
Fredric
Katherine
Jillian
Sella
Graehme
Dasi
Maxie
Anselma
Dixie
Carlota
Romeo
Brnaby
Chen
Ced
Britni
Leonore
Rachel
Haley
Dorotea
Filia
Reeta
Cynde
Willy
Clara
Joyann
Debbie
Krista
Torin
Mandel
Ravid
Madelene
Bondon
Nita
Josephina
Luce
Arlette
Kelvin
Fenelia
Palmer
Enrico
Darlleen
Ossie
Simmonds
Wesley
Vivien
Rusty
Aron
Melesa
Klarika
Milton
Bekki
Opaline
Anet
Monty
Daloris
Amos
Gianni
Leann
Euell
Alvin
Harman
Chrisse
Roosevelt
`,

	"emails": `
cbewlay0@nhs.uk
asprankling1@slate.com
mpawlicki2@pagesperso-orange.fr
terrington3@cam.ac.uk
hlitherland4@independent.co.uk
gcoxhell5@tripadvisor.com
khutchin6@sbwire.com
isturdgess7@xrea.com
aandreolli8@ucla.edu
clouedey9@instagram.com
rlanstona@usatoday.com
gbalentyneb@columbia.edu
tbufferyc@newsvine.com
spennid@live.com
srobline@army.mil
aduetschef@usgs.gov
slecornug@addtoany.com
akermitth@jalbum.net
doconnollyi@ow.ly
kmccloyj@globo.com
tgosneyek@netvibes.com
fkalbl@elpais.com
jpinckardm@discovery.com
cskerrittn@unesco.org
hworstallo@flickr.com
acronkshawp@youtube.com
riacovoq@artisteer.com
mmcloneyr@independent.co.uk
kgolightlys@mashable.com
acastangiat@patch.com
kgermonu@plala.or.jp
scremenv@wired.com
liorizziw@elegantthemes.com
djancikx@sciencedaily.com
wrapelliy@cafepress.com
wandrockz@miitbeian.gov.cn
rfendlen10@cdbaby.com
lchastan11@networksolutions.com
eicom12@ihg.com
amacklin13@addthis.com
adarwin14@altervista.org
mmacenzy15@desdev.cn
mudden16@com.com
rbiagi17@noaa.gov
jcurtoys18@scribd.com
dskurm19@microsoft.com
cjevons1a@devhub.com
cdecastri1b@google.de
glatchmore1c@histats.com
tpedwell1d@unesco.org
afabri1e@vinaora.com
cpeever1f@over-blog.com
mmurcutt1g@forbes.com
gphilliphs1h@sphinn.com
rclamp1i@1688.com
mcollibear1j@cargocollective.com
vhaslam1k@liveinternet.ru
gmeron1l@hp.com
csaward1m@google.de
asandham1n@timesonline.co.uk
cvelasquez1o@unblog.fr
abailles1p@reddit.com
flukasik1q@bing.com
kpiffe1r@utexas.edu
pbalsom1s@instagram.com
cormerod1t@clickbank.net
gmaclleese1u@issuu.com
edocksey1v@webmd.com
scammiemile1w@sphinn.com
hcolthurst1x@over-blog.com
aclute1y@about.me
dhassall1z@blogspot.com
msalvidge20@histats.com
bcrix21@wordpress.com
ocarwithim22@forbes.com
ithresher23@dedecms.com
rtupper24@nationalgeographic.com
cmcorkill25@tmall.com
kgantlett26@bbc.co.uk
vborrott27@wordpress.org
eiacobini28@fda.gov
nbuxsey29@deliciousdays.com
lclewley2a@wikipedia.org
klathe2b@virginia.edu
lcuttelar2c@webnode.com
cseiller2d@huffingtonpost.com
vlosselyong2e@newyorker.com
hfullun2f@moonfruit.com
lirlam2g@mayoclinic.com
gbindin2h@tripod.com
ecolloff2i@oaic.gov.au
ssebrook2j@rediff.com
cocorrigane2k@jimdo.com
jpurdon2l@ezinearticles.com
cfrazer2m@google.com.hk
mbellas2n@bbc.co.uk
ccallway2o@discuz.net
slebbern2p@cam.ac.uk
hdodimead2q@slashdot.org
gfinding2r@irs.gov
ehovert0@wikimedia.org
lpidon1@alibaba.com
rgerrad2@ucoz.com
ftowey3@mac.com
vfranck4@deliciousdays.com
yapthorpe5@istockphoto.com
dgull6@state.gov
ikitley7@unicef.org
dkristoffersson8@gnu.org
asaggers9@cdbaby.com
tnutona@quantcast.com
jdubbleb@google.cn
grantoullc@tiny.cc
ablenkirond@symantec.com
nsmailse@webs.com
ekeeffef@aol.com
lmccracheng@etsy.com
gfawloeh@parallels.com
lpowderhami@biglobe.ne.jp
tgillicej@php.net
wbeecraftk@unblog.fr
mbleil@google.it
sbrahmm@webnode.com
eivanishinn@de.vu
mmorato@ihg.com
gcotmorep@hhs.gov
fmeekq@seesaa.net
dshemmansr@nsw.gov.au
lmatusovs@loc.gov
jcramert@surveymonkey.com
lpiotrkowskiu@cbsnews.com
cculleyv@behance.net
jperhamw@typepad.com
msandilandx@hexun.com
sshinglesy@jimdo.com
fcommuzzoz@ow.ly
daxelbee10@umn.edu
ebudgett11@oakley.com
bhallock12@census.gov
ecannon13@slate.com
cgauthorpp14@discuz.net
rantonognoli15@wsj.com
jmarley16@sina.com.cn
lcrutch17@scientificamerican.com
ivina18@exblog.jp
frogeron19@github.io
bgoning1a@sitemeter.com
rcranstone1b@yellowbook.com
kwrathmall1c@infoseek.co.jp
icullotey1d@bloglovin.com
cdaughtrey1e@ebay.co.uk
upolkinhorn1f@google.com.br
pgreschik1g@creativecommons.org
ggillease1h@wikispaces.com
wgurnett1i@storify.com
adyers1j@constantcontact.com
jcusick1k@bloomberg.com
jmacoun1l@storify.com
sbolletti1m@ucla.edu
mbenoit1n@mediafire.com
jmadine1o@theguardian.com
jsaxelby1p@berkeley.edu
sisland1q@yahoo.com
bcasemore1r@twitter.com
chexum1s@bluehost.com
eburkwood1t@yelp.com
dtorrese1u@amazon.co.uk
tbampkin1v@spotify.com
ahully1w@cbsnews.com
lmortell1x@joomla.org
aspare1y@stanford.edu
lelvins1z@photobucket.com
gulyatt20@skyrock.com
obeatson21@vistaprint.com
pswan22@mapquest.com
zblanchette23@seesaa.net
vbleasby24@nasa.gov
breay25@ihg.com
wcamerello26@weibo.com
baxon27@quantcast.com
kboskell28@cloudflare.com
bsonschein29@youtube.com
gparkes2a@netvibes.com
gguittet2b@google.de
tdineges2c@sbwire.com
twarbeys2d@dot.gov
rbaggelley2e@buzzfeed.com
adarrington2f@sohu.com
wgrinsted2g@addtoany.com
nrandales2h@xing.com
cbeauman2i@cdbaby.com
rcawse2j@g.co
dhatwells2k@newsvine.com
ledgson2l@taobao.com
fclemo2m@homestead.com
mfirsby2n@lulu.com
soakly2o@bbb.org
apimlott2p@dedecms.com
amacintosh2q@google.com
tfurbank2r@yellowpages.com
egrosier0@utexas.edu
bbrimson1@networksolutions.com
dkenford2@nyu.edu
estigers3@dailymail.co.uk
cleathard4@umich.edu
gpym5@answers.com
cellesworthe6@springer.com
kmishaw7@omniture.com
hlabbati8@psu.edu
fwagnerin9@amazon.co.jp
chammelberga@imgur.com
mpettetb@adobe.com
sbroadbridgec@phpbb.com
gtometd@china.com.cn
jfrancklyne@rambler.ru
tcusackf@cisco.com
tcarryerg@sciencedirect.com
fivimeyh@stanford.edu
ichillistonei@de.vu
bkloisnerj@baidu.com
kgarrattleyk@flickr.com
krheaml@biglobe.ne.jp
ljollimanm@e-recht24.de
cbartheln@posterous.com
flequesneo@stanford.edu
vmctrustiep@blogs.com
fbohlensq@deviantart.com
locloneyr@hexun.com
ymccomass@un.org
vboamet@ucoz.ru
djimmesu@cdbaby.com
bflancinbaumv@va.gov
lcustancew@intel.com
gfetterx@google.cn
tduffelly@amazon.com
wfinicjz@jiathis.com
cvanmerwe10@spotify.com
mfoukx11@wp.com
gpressman12@upenn.edu
cjunkison13@biglobe.ne.jp
sacory14@japanpost.jp
ado15@springer.com
ecox16@hibu.com
smungham17@odnoklassniki.ru
jstilly18@i2i.jp
mguerola19@smugmug.com
fdudin1a@yahoo.com
asymms1b@mapquest.com
slebretondelavieuville1c@live.com
mlytle1d@purevolume.com
fstolting1e@vinaora.com
jstocking1f@auda.org.au
knowland1g@blogspot.com
fcasperri1h@clickbank.net
mlovatt1i@youku.com
kshutler1j@yellowbook.com
olambourne1k@vk.com
thothersall1l@va.gov
fdollman1m@tumblr.com
jadnam1n@mashable.com
dsaintpierre1o@uol.com.br
sglazer1p@fema.gov
rcurtis1q@mozilla.com
gwoolis1r@ocn.ne.jp
asimionato1s@reddit.com
wpattesall1t@moonfruit.com
ldebruyne1u@tripadvisor.com
icheal1v@rambler.ru
tadriano1w@chronoengine.com
amoncreif1x@abc.net.au
mburnapp1y@cnbc.com
gmacgiany1z@nationalgeographic.com
kyablsley20@techcrunch.com
rkarolyi21@virginia.edu
jclimance22@netlog.com
kguiduzzi23@mac.com
aottery24@booking.com
asiddele25@yandex.ru
soriordan26@amazon.co.uk
sriddles27@senate.gov
knerney28@last.fm
rsafell29@craigslist.org
rcastledine2a@marketwatch.com
rdrife2b@wiley.com
mshingfield2c@quantcast.com
dduprey2d@ted.com
jmacphee2e@nifty.com
pcomo2f@nationalgeographic.com
jsouthey2g@wikia.com
ksheppard2h@php.net
gdavy2i@hostgator.com
nmccowen2j@google.com.hk
abullus2k@naver.com
abasketter2l@npr.org
ekrelle2m@ow.ly
umatyashev2n@sohu.com
fbrach2o@netscape.com
wmirfin2p@github.io
mbilham2q@parallels.com
trhodus2r@tinypic.com
`,
}
