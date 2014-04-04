package main

import (
    "fmt"
    "flag"
    "strings"
    "regexp"
    "os"
    "os/exec"
    "io/ioutil"
)

var old_str = ""
var new_str = ""
var xml_file = ""
var res_path = "./"  //  resource folder
var overlay_res_path = "" // overlay resource folder

var is_overlay_mode = false

var manual = `Manual:
replace_android_str_res is a tool to replace strings in xml files(<string>xxxx</string>) with new strings under Android app resource path(xx/res/).
It will search / replace all files for multi-language resources under /res. Ex: values-zh-rCN/strings.xml...
There're 2 modes:

-------------------
1. Overlay Mode.
-------------------
It will search <string>xxx</string> in original xml files and only copy the item which contains the string to be replaced to new xml files under overlay folder.
 
Usage:
replace_android_str_res -s <old_string> -n <new_string> -f <xml file> -r <Android app resource path> -o <Android app resource overlay path>

Ex:
replace_android_str_res -s "Miracast" -n "Wireless Display" -f "miracast_strings.xml" -r ~/proj/packages/app/settings/res -o ~/proj/overlay/packages/app/settings/res

-------------------
2. Overwrite Mode.
-------------------
It will find <string>xxx</string> item which contains the string to be replaced then overwrite the original xml files.
      
Usage:
replace_android_str_res -s <old_string> -n <new_string> -f <xml file> -r <Android app resource path>

Ex:
replace_android_str_res -s "Miracast" -n "Wireless Display" -f "miracast_strings.xml" -r ~/proj/packages/app/settings/res`

func main() {
    flag.StringVar(&old_str, "s", "", "old string")
    flag.StringVar(&new_str, "n", "", "new string")
    flag.StringVar(&xml_file, "f", "", "xml file name without path. Ex string.xml")
    flag.StringVar(&res_path, "r", "./", "Android app resource folder path. Ex: ~/proj/packages/apps/settings/res")
    flag.StringVar(&overlay_res_path, "o", "", "Android app overlay resource folder path to output new xml files. Ex: ~/proj/overlay/packages/apps/settings/res")

    flag.Parse()

    fmt.Println("old_str: " + old_str)
    fmt.Println("new_str: " + new_str)
    fmt.Println("xml_file: " + xml_file)
    fmt.Println("res_path: " + res_path)
    fmt.Println("overlay_res_path: " + overlay_res_path)

    if len(old_str) == 0 || len(xml_file) == 0 {
        fmt.Println(manual)
        return
    }

    if len(overlay_res_path) == 0 {
        is_overlay_mode = false
    } else {
        is_overlay_mode = true
    }

    fmt.Printf("is_overlay_mode: %v\n", is_overlay_mode)

    pattern := fmt.Sprintf(`.*/?res/(?P<folder>.*)/%s:\d.*:\s*(?P<item><string .*>.*</string>)`, xml_file)

    // Step 1. Use find and grep to search "Miracast" related strings in miracast_strings.xml.

    // use this method, we'll get 'exit status 123' error by xargs, see:
    // http://lists.gnu.org/archive/html/help-gnu-emacs/2013-04/msg00340.html
    cmd := fmt.Sprintf("find %s -name '%s' | xargs grep -e '<string .*>.*%s.*</string>' -rn --color", res_path, xml_file, old_str)
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

    }else {
        fmt.Println("\nNot found")
    }

    // Step 2. Use regular expression to parse language name, region and replace string items in the output.

    re := regexp.MustCompile(pattern)
    matchedItems := re.FindAllStringSubmatch(string(out), -1)

    stringMap := make(map[string][]string)  // key: values folder name(Ex: values, values-zh-rCN). value: new string slice contains "Miracast"(Ex: <string name="xxxx">Miracast</string>.

    // for overwirte mode
    old_stringMap := make(map[string][]string)  // key: value folder name(Ex: values, values-zh-rCn). value: old string item with the same index as new string slice in stringMap.

    for i := 0; i < len(matchedItems); i++ {
        folder := matchedItems[i][1]
        item := matchedItems[i][2]

        new_item := strings.Replace(item, old_str, new_str, -1)  // replace string
        stringMap[folder] = append(stringMap[folder], new_item)
        if !is_overlay_mode {
            old_stringMap[folder] = append(old_stringMap[folder], item)  // record old string
        }
    }

    // Step 3. Create new miracast_strings.xml and values folder under Setttings overlay folder.
    for k, v := range stringMap {
        if is_overlay_mode {
            // overlay mode
            //fmt.Println(k)
            folder := overlay_res_path + "/" + k
            //fmt.Println(folder)
            if err := os.MkdirAll(folder, os.ModeDir | os.ModePerm); err != nil {
                fmt.Println(err)
                return
            }

            file_name := folder + "/" + xml_file
            f, err := os.Create(file_name)
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
            file_name := res_path + "/" + k + "/" + xml_file
            //fmt.Println(file_name)
            buf, err := ioutil.ReadFile(file_name)
            if err != nil {
                fmt.Println(err)
                return
            }

            if len(buf) == 0 {
                fmt.Println("empty file: " + file_name)
                return
            }

            s := string(buf)
            for i := 0; i < len(v); i++ {
                s = strings.Replace(s, old_stringMap[k][i], v[i], -1)  // replace old item with new item
            }
            if err := ioutil.WriteFile(file_name, []byte(s), os.ModePerm); err != nil {
                fmt.Println(err)
                return
            }
        }
    }

    fmt.Printf("\n---------------------------\nTotal: %v files replaced.\n", len(stringMap))

    return
}
