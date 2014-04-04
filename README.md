
# Manual

`replace_android_str_res` is a tool to replace strings in xml files(<string>xxxx</string>) with new strings under Android app resource path(xx/res/).

It will search / replace all files for multi-language resources under `yourapp/res`. Ex: `values-zh-rCN/strings.xml`...  

There're 2 modes:

1. Overlay Mode.
-------------------
It will search <string>xxx</string> in original xml files and only copy the item which contains the string to be replaced to new xml files under overlay folder.

Usage:  
`replace_android_str_res -s <old_string> -n <new_string> -f <xml file> -r <Android app resource path> -o <Android app resource overlay path>`

Ex:  
`replace_android_str_res -s "Miracast" -n "Wireless Display" -f "miracast_strings.xml" -r ~/proj/packages/app/settings/res -o ~/proj/overlay/packages/app/settings/res`

2. Overwrite Mode.
-------------------
It will find <string>xxx</string> item which contains the string to be replaced then overwrite the original xml files.

Usage:  
`replace_android_str_res -s <old_string> -n <new_string> -f <xml file> -r <Android app resource path>`

Ex:  
`replace_android_str_res -s "Miracast" -n "Wireless Display" -f "miracast_strings.xml" -r ~/proj/packages/app/settings/res``
