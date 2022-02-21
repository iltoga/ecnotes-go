# EcNotes (encrypted notes)
Multi-platform gui app to store and manage encrypted notes. It can be used to store sensitive information such as passwords and crypto keys

EcNotes starts as a hobby project to get away some 'rust' from my golang programming and develop something useful (at least for me).
Through the years I've always been struggling to find a simple tool to store my passwords and other sensitive information and that meets the following requirements:

- must be multiplatform: must run on desktop and mobile as well
- must be locally installed as a GUI application: I don't want to rely on some third party-internet software
- must be able to sync **encrypted** data to some cloud storage/db services as an optional feature: this is required if I want to use the app on different systems/devices and retain all my data
- must give me the ownership of my data: I want to be able to generate my own encryption key/s locally, save them where I want and storing or syncing with cloud services only encrypted content. meaning, the only time where my data are in clear text is inside the application and only for the piece of data I am actually accessing (single note)
- would be nice if it allows to choose between different encryption algorithms
- would be nice if it allows to manage external (public) keys: this would allow to exchange encrypted notes/messages with other people (you know.. real e2e encryption, without having to trust third parties ;) )


