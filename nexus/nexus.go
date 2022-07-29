package nexus

import (
	"encoding/json"
	"errors"
	"fmt"
	converter "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/frustra/bbcode"
	"github.com/kiamev/moogle-mod-manager/config"
	"github.com/kiamev/moogle-mod-manager/mods"
	"io"
	"net/http"
	"strings"
)

type NexusGame string
type NexusGameID int

const (
	FFI   NexusGame = "finalfantasypixelremaster"
	FFII  NexusGame = "finalfantasy2pixelremaster"
	FFIII NexusGame = "finalfantasy3pixelremaster"
	FFIV  NexusGame = "finalfantasy4pixelremaster"
	FFV   NexusGame = "finalfantasy5pixelremaster"
	FFVI  NexusGame = "finalfantasy6pixelremaster"

	IdFFI   NexusGameID = 3934
	IdFFII  NexusGameID = 3958
	IdFFIII NexusGameID = 3942
	IdFFIV  NexusGameID = 4022
	IdFFV   NexusGameID = 4137
	IdFFVI  NexusGameID = 4335

	nexusApiModUrl         = "https://api.nexusmods.com/v1/games/%s/mods/%s.json"
	nexusApiModDlUrl       = "https://api.nexusmods.com/v1/games/%s/mods/%s/files.json%s"
	nexusApiModDlUrlSuffix = "?category=main%2Cupdate%2Coptional%2Cmiscellaneous"
	nexusUrl               = "https://www.nexusmods.com/%s/mods/%d"

	nexusUsersApiUrl = "https://users.nexusmods.com/oauth/token"

	// file_id, NexusGameID
	nexusFileDownload = "https://www.nexusmods.com/Core/Libs/Common/Widgets/DownloadPopUp?id=%d&game_id=%v"
)

func IsNexus(url string) bool {
	return strings.Index(url, "nexusmods.com") >= 0
}

func GetModFromNexus(url string) (mod *mods.Mod, err error) {
	var (
		sp      = strings.Split(url, "/")
		nexusID string
		modID   string
		b       []byte
		nMod    nexusMod
		nDls    fileParent
	)
	for i, s := range sp {
		if s == "mods" {
			if i > 0 && i < len(sp)-1 {
				nexusID = sp[i-1]
				modID = sp[i+1]
				break
			}
		}
	}
	if nexusID == "" && modID == "" {
		err = fmt.Errorf("could not get Game and Mod ID from %s", url)
		return
	}
	if b, err = sendRequest(fmt.Sprintf(nexusApiModUrl, nexusID, modID)); err != nil {
		return
	}
	if err = json.Unmarshal(b, &nMod); err != nil {
		return
	}

	if b, err = sendRequest(fmt.Sprintf(nexusApiModDlUrl, nexusID, modID, nexusApiModDlUrlSuffix)); err != nil {
		return
	}
	if err = json.Unmarshal(b, &nDls); err != nil {
		return
	}
	return toMod(nMod, nDls.Files)
}

func sendRequest(url string) (response []byte, err error) {
	var (
		apiKey = config.GetSecrets().NexusApiKey
		req    *http.Request
		resp   *http.Response
	)
	if apiKey == "" {
		err = errors.New("no Nexus Api Key set. Please to to File->Secrets")
		return
	}
	if req, err = http.NewRequest(http.MethodGet, url, nil); err != nil {
		err = fmt.Errorf("failed to create request to validate user with nexus %s: %v", url, err)
		return
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("apikey", apiKey)

	if resp, err = (&http.Client{}).Do(req); err != nil {
		err = fmt.Errorf("failed to make request to %s: %v", url, err)
		return
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	if code < 200 && code >= 300 {
		err = fmt.Errorf("received code [%d] from call to [%s]", code, url)
		return
	}

	if response, err = io.ReadAll(resp.Body); err != nil {
		err = fmt.Errorf("failed to read response's body for %s: %v", url, err)
	}
	return
}

func toMod(n nexusMod, dls []NexusFile) (mod *mods.Mod, err error) {
	modID := fmt.Sprintf("%d", n.ModID)
	var nexusGameID NexusGameID
	mod = &mods.Mod{
		ID:           modID,
		Name:         n.Name,
		Author:       n.Author,
		AuthorLink:   n.AuthorLink,
		Category:     fmt.Sprintf("%d", n.CategoryID),
		ReleaseDate:  n.CreatedTime.Format("Jan 2, 2006"),
		ReleaseNotes: "",
		Link:         fmt.Sprintf(nexusUrl, n.Game, n.ModID),
		Preview: &mods.Preview{
			Url:   nil,
			Local: nil,
			Size:  mods.Size{},
		},
		ModKind: &mods.ModKind{
			Kind: mods.Nexus,
			Nexus: &mods.NexusModKind{
				ID: modID,
			},
		},
		Games:          []*mods.Game{&mods.Game{}},
		Downloadables:  make([]*mods.Download, len(dls)),
		DonationLinks:  nil,
		AlwaysDownload: nil,
		Configurations: nil,
	}
	compiler := bbcode.NewCompiler(true, true)
	c := converter.NewConverter("", true, nil)
	if mod.Description, err = c.ConvertString(compiler.Compile(n.Description)); err != nil {
		mod.Description = n.Description
		err = nil
	}
	switch n.Game {
	case FFI:
		mod.Games[0].Name = config.FfPrI
		nexusGameID = IdFFI
	case FFII:
		mod.Games[0].Name = config.FfPrII
		nexusGameID = IdFFII
	case FFIII:
		mod.Games[0].Name = config.FfPrIII
		nexusGameID = IdFFIII
	case FFIV:
		mod.Games[0].Name = config.FfPrIV
		nexusGameID = IdFFIV
	case FFV:
		mod.Games[0].Name = config.FfPrV
		nexusGameID = IdFFV
	case FFVI:
		mod.Games[0].Name = config.FfPrVI
		nexusGameID = IdFFVI
	default:
		err = fmt.Errorf("unsupported game %s", n.Game)
		return
	}

	for i, d := range dls {
		mod.Downloadables[i] = &mods.Download{
			Name:    d.Name,
			Version: d.Version,
			Nexus: &mods.NexusDownloadable{
				FileID:   d.FileID,
				FileName: d.FileName,
			},
			DownloadedLoc: fmt.Sprintf(nexusFileDownload, d.FileID, nexusGameID),
		}
	}
	return
}

/*
function convert() {
  var left = document.getElementById('left_ta');
  var right = document.getElementById('right_ta');

  var left_value = left.value;

  //preprocessing for tf2toolbox BBCode
  if(left_value.search(/TF2Toolbox/gmi) != -1) {
    left_value = left_value
    .replace(/(\(List generated at .+?\[\/URL\]\))((?:.|\n)+)/gmi, '$2\n\n\n$1') //Move TF2Toolbox link to bottom
    .replace('(List generated at', '(List generated from')
    .replace(/[^\S\n]+\(List/gmi,'(List')
    .replace(/\[b\]\[u\](.+?)\[\/u\]\[\/b\]/gmi,'[b]$1[/b]\n') //Fix double emphasized titles
    .replace(/(\n)\[\*\]\[b\](.+?)\[\/b\]/gmi, '$1\[\*\] $2');
  }

  //general BBcode conversion
  left_value = left_value
    .replace(/\[b\]((?:.|\n)+?)\[\/b\]/gmi, '**$1**') //bold; replace [b] $1 [/b] with ** $1 **
    .replace(/\[\i\]((?:.|\n)+?)\[\/\i\]/gmi, '*$1*')  //italics; replace [i] $1 [/u] with * $1 *
    .replace(/\[\u\]((?:.|\n)+?)\[\/\u\]/gmi, '$1')  //remove underline;
    .replace(/\[s\]((?:.|\n)+?)\[\/s\]/gmi, '~~ $1~~') //strikethrough; replace [s] $1 [/s] with ~~ $1 ~~
    .replace(/\[center\]((?:.|\n)+?)\[\/center\]/gmi, '$1') //remove center;
    .replace(/\[quote\=.+?\]((?:.|\n)+?)\[\/quote\]/gmi, '$1') //remove [quote=] tags
    .replace(/\[size\=.+?\]((?:.|\n)+?)\[\/size\]/gmi, '## $1') //Size [size=] tags
    .replace(/\[color\=.+?\]((?:.|\n)+?)\[\/color\]/gmi, '$1') //remove [color] tags
    .replace(/\[list\=1\]((?:.|\n)+?)\[\/list\]/gmi, function (match, p1, offset, string) {return p1.replace(/\[\*\]/gmi, '1. ');})
    .replace(/(\n)\[\*\]/gmi, '$1* ') //lists; replcae lists with + unordered lists.
    .replace(/\[\/*list\]/gmi, '')
    .replace(/\[img\]((?:.|\n)+?)\[\/img\]/gmi,'![$1]($1)')
    .replace(/\[url=(.+?)\]((?:.|\n)+?)\[\/url\]/gmi,'[$2]($1)')
    .replace(/\[code\](.*?)\[\/code\]/gmi, '`$1`')
    .replace(/\[code\]((?:.|\n)+?)\[\/code\]/gmi, function (match, p1, offset, string) {return p1.replace(/^/gmi, '    ');})
    .replace(/\[php\](.*?)\[\/php\]/gmi, '`$1`')
    .replace(/\[php\]((?:.|\n)+?)\[\/php\]/gmi, function (match, p1, offset, string) {return p1.replace(/^/gmi, '    ');})
    .replace(/\[pawn\](.*?)\[\/pawn\]/gmi, '`$1`')
    .replace(/\[pawn\]((?:.|\n)+?)\[\/pawn\]/gmi, function (match, p1, offset, string) {return p1.replace(/^/gmi, '    ');});

  //post processing for tf2toolbox BBCode
  if(left_value.search(/TF2Toolbox/gmi) != -1) {
    left_value = left_value
    .replace('/bbcode_lookup.php))', '/bbcode_lookup.php) and converted to /r/tf2trade ready Markdown by Dum\'s [converter](http://jondum.github.com/BBCode-To-Markdown-Converter/)).') //add a linkback
    .replace(/\*\*.+?\*\*[\s]+?None[\s]{2}/gmi, ''); //remove empty sections

  }

  right.value = left_value;

}
*/
