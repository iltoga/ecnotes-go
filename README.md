
# EcNotes (encrypted notes) 

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

<img src="https://github.com/iltoga/ecnotes-go/blob/main/Icon.png" width="50" height="50">

Multi-platform gui app to store and manage encrypted notes. It can be used to store sensitive information such as passwords and crypto keys

EcNotes starts as a hobby project to get away some 'rust' from my golang programming and develop something useful (at least for me).
Through the years I've always been struggling to find a simple tool to store my passwords and other sensitive information and that meets the following requirements:

- must be multiplatform: must run on desktop and mobile as well
- must be locally installed as a GUI application: I don't want to rely on some third party-internet software
- must be able to sync **encrypted** data to some cloud storage/db services as an optional feature: this is required if I want to use the app on different systems/devices and retain all my data
- must give me the ownership of my data: I want to be able to generate my own encryption key/s locally, save them where I want and storing or syncing with cloud services only encrypted content. meaning, the only time where my data are in clear text is inside the application and only for the piece of data I am actually accessing (single note)
- would be nice if it allows to choose between different encryption algorithms
- would be nice if it allows to manage external (public) keys: this would allow to exchange encrypted notes/messages with other people (you know.. real e2e encryption, without having to trust third parties ;) )

### INSTALLATION
At the moment the only way to install EcNotes is to compile it (maybe in future I will provide executables for different platforms..).
You will have to first install these dependencies:
- go (tested on versions >= 1.17)
- git
- fyne commandline utility: instructions [here](https://developer.fyne.io/started/packaging)

#### PACKAGE ECNOTES (DISTRIBUTABLE)
To package EcNotes (build the executable and installer package), go to the gooroot folder $GOROOT/src/github (usually ~/go/src/github) and create github dir if not present:
```
git clone git@github.com:iltoga/ecnotes-go.git
cd ecnotes-go

# native (for your running OS/platform):
fyne package -icon Icon.png

# other os (read instructions on https://developer.fyne.io/started/packaging to install relative OS toolchains), eg:

fyne package -os linux -icon Icon.png
fyne package -os windows -icon Icon.png
```
Then you should be able to install the app with your OS packaging system.

#### JUST BUILD AND INSTALL LOCALLY
Go to the gooroot folder $GOROOT/src/github (usually ~/go/src/github) and create github dir if not present:
```
git clone git@github.com:iltoga/ecnotes-go.git
cd ecnotes-go

fyne install -icon Icon.png
```

\* alternatively you can just use ```go run .```  or ``` go build .``` to run or build the executable, without installing it into some other directory



### EXTERNAL PROVIDERS
You can use these providers to extend functionalities of EcNotes:
- Google
- TODO: add others...

### Google
With this provider you can sync (two-way) your ecnotes to a google sheet on your google account.
This enables:
- Database synchronization between multiple instances of EcNotes (eg. one on a linux desktop and another on an Android phone) via google sheet
- Cloud backup service via google sheet 

#### Setting up your Google account
To set up this provider you must first configure an app and service account* using Google Developer Console:
https://console.developers.google.com

This article has an example on how to do it 
- How to setup google api app and service account: "Authenticating with Google Sheets API" paragraph
- How to create and share a google sheet with this service account: "Share your spreadsheet with" sub paragraph

https://blog.coupler.io/how-to-use-google-sheets-as-database/#Exportimport_data_automatically_using_the_Google_Sheets_API

\* for now we only support authentication via 'service account' credentials, which doesn't require to authenticate your google app via web (oauth2) and requires to share your google sheet with the service account email that will be created during the procedure described in the article.

#### Format the google sheet
Once you have set up your Google account and created and shared your google sheet, you have to format it by adding these column headers in the first row:

| ID | Title | Content | Hidden | Encrypted | CreatedAt|UpdatedAt |
|----|-------|---------|--------|-----------|----------|----------|

#### Configuring EcNotes with google account sync
If you have followed the article and created the google account service, you've been asked to download the json file with the credentials to your computer. 
Now run EcNotes at least once, so that creates the configuration directories and copy or move this file to your home directory (on linux 'echo #HOME' from a terminal to see where it is). Eg: 
`mv whatever_credentials.json #HOME/.config/ecnotes/providers/google/cred_serviceaccount.json`
<i>Please note that the file name (`cred_serviceaccount.json`) is important.<\i>

Last but not least is to get your google sheet ID and add it to EcNotes configuration:
- copy the id from the sheet url (it looks something like this `1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms`)
- Add this line into your `config.toml` file (that should be in `$HOME/.config/ecnotes/resources/`):
	```toml
	google_sheet_id = "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"
	```
	don't forget to change the id with yours
