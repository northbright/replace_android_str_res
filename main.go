package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	oldStr         = ""
	newStr         = ""
	xmlFile        = ""
	resPath        = "./" //  resource folder
	overlayResPath = ""   // overlay resource folder
	isOverlayMode  = false
	manual         = `Manual:
replace_android_str_res is a tool to replace strings in xml files(<string>xxxx</string>) with new strings under Android app resource path(xx/res/).
It will search / replace all files for multi-language resources under /res. Ex: values-zh-rCN/strings.xml...
There're 2 modes:

-------------------
1. Overlay Mode.
-------------------
It will search <string>xxx</string> in original xml files and only copy the item which contains the string to be replaced to new xml files under overlay folder.
 
Usage:
replace_android_str_res -s <oldString> -n <newString> -f <xml file> -r <Android app resource path> -o <Android app resource overlay path>

Ex:
replace_android_str_res -s "Miracast" -n "Wireless Display" -f "miracast_strings.xml" -r ~/proj/packages/app/settings/res -o ~/proj/overlay/packages/app/settings/res

-------------------
2. Overwrite Mode.
-------------------
It will find <string>xxx</string> item which contains the string to be replaced then overwrite the original xml files.
      
Usage:
replace_android_str_res -s <oldString> -n <newString> -f <xml file> -r <Android app resource path>

Ex:
replace_android_str_res -s "Miracast" -n "Wireless Display" -f "miracast_strings.xml" -r ~/proj/packages/app/settings/res`
)

func main() {
	flag.StringVar(&oldStr, "s", "", "old string")
	flag.StringVar(&newStr, "n", "", "new string")
	flag.StringVar(&xmlFile, "f", "", "xml file name without path. Ex string.xml")
	flag.StringVar(&resPath, "r", "./", "Android app resource folder path. Ex: ~/proj/packages/apps/settings/res")
	flag.StringVar(&overlayResPath, "o", "", "Android app overlay resource folder path to output new xml files. Ex: ~/proj/overlay/packages/apps/settings/res")

	flag.Parse()

	fmt.Println("oldStr: " + oldStr)
	fmt.Println("newStr: " + newStr)
	fmt.Println("xmlFile: " + xmlFile)
	fmt.Println("resPath: " + resPath)
	fmt.Println("overlayResPath: " + overlayResPath)

	if len(oldStr) == 0 || len(xmlFile) == 0 {
		fmt.Println(manual)
		return
	}

	if len(overlayResPath) == 0 {
		isOverlayMode = false
	} else {
		isOverlayMode = true
	}

	fmt.Printf("isOverlayMode: %v\n", isOverlayMode)

	pattern := fmt.Sprintf(`.*/?res/(?P<folder>.*)/%s:\d.*:\s*(?P<item><string .*>.*</string>)`, xmlFile)

	// Step 1. Use find and grep to search "Miracast" related strings in miracast_strings.xml.

	// use this method, we'll get 'exit status 123' error by xargs, see:
	// http://lists.gnu.org/archive/html/help-gnu-emacs/2013-04/msg00340.html
	cmd := fmt.Sprintf("find %s -name '%s' | xargs grep -e '<string .*>.*%s.*</string>' -rn --color", resPath, xmlFile, oldStr)
	fmt.Println(cmd)

	// use this method, no 'exit status 123' error code. But it's much slower.
	//cmd := fmt.Sprintf("find %s -name '%s' -type f -exec grep '<string .*>.*%s.*</string>' -rn --color {} \\;", path, ,"miracast_strings.xml", "Miracast")
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		// xargs will return 123 when grep does not match
		if err.Error() != "exit status 123" {
			fmt.Println(err)
		}
	}
	if len(out) != 0 {
		//fmt.Printf("%s", out)
		fmt.Println("\nFound:")
		fmt.Printf("%s", out)
		//foundStrings = append(foundStrings, string(s))

	} else {
		fmt.Println("\nNot found")
	}

	// Step 2. Use regular expression to parse language name, region and replace string items in the output.

	re := regexp.MustCompile(pattern)
	matchedItems := re.FindAllStringSubmatch(string(out), -1)

	stringMap := make(map[string][]string) // key: values folder name(Ex: values, values-zh-rCN). value: new string slice contains "Miracast"(Ex: <string name="xxxx">Miracast</string>.

	// for overwirte mode
	oldStringMap := make(map[string][]string) // key: value folder name(Ex: values, values-zh-rCn). value: old string item with the same index as new string slice in stringMap.

	for i := 0; i < len(matchedItems); i++ {
		folder := matchedItems[i][1]
		item := matchedItems[i][2]

		newItem := strings.Replace(item, oldStr, newStr, -1) // replace string
		stringMap[folder] = append(stringMap[folder], newItem)
		if !isOverlayMode {
			oldStringMap[folder] = append(oldStringMap[folder], item) // record old string
		}
	}

	// Step 3. Create new miracast_strings.xml and values folder under Setttings overlay folder.
	for k, v := range stringMap {
		if isOverlayMode {
			// overlay mode
			//fmt.Println(k)
			folder := overlayResPath + "/" + k
			//fmt.Println(folder)
			if err := os.MkdirAll(folder, os.ModeDir|os.ModePerm); err != nil {
				fmt.Println(err)
				return
			}

			fileName := folder + "/" + xmlFile
			f, err := os.Create(fileName)
			defer f.Close()
			if err != nil {
				fmt.Println(err)
				return
			}

			s := "<?xml version=\"1.0\" encoding=\"utf-8\"?>\n"
			s += "<resources>\n"
			for i := 0; i < len(v); i++ {
				//fmt.Println(v[i])
				s += "    " + v[i] + "\n"
			}
			s += "</resources>\n"
			fmt.Println("==================\n" + s)

			if _, err := f.WriteString(s); err != nil {
				fmt.Println(err)
			}
		} else {
			// overwrite mode
			fileName := resPath + "/" + k + "/" + xmlFile
			//fmt.Println(fileName)
			buf, err := ioutil.ReadFile(fileName)
			if err != nil {
				fmt.Println(err)
				return
			}

			if len(buf) == 0 {
				fmt.Println("empty file: " + fileName)
				return
			}

			s := string(buf)
			for i := 0; i < len(v); i++ {
				s = strings.Replace(s, oldStringMap[k][i], v[i], -1) // replace old item with new item
			}
			if err := ioutil.WriteFile(fileName, []byte(s), os.ModePerm); err != nil {
				fmt.Println(err)
				return
			}
		}
	}

	fmt.Printf("\n---------------------------\nTotal: %v files replaced.\n", len(stringMap))

	return
}
