How to properly set up Sweetiebot for use on your server

**1.** Install [Golang](https://golang.org/dl/) and [MariaDB](https://downloads.mariadb.org/) to your computer. Take note of your credentials for MariaDB, as you'll need to know that for later.

**2.** Verify that Go was properly installed to your PATH variable by typing "go ver" in your terminal / command prompt. If you aren't prompted with something Go related, restart your computer and try again.

**3.** Clone the repository in it's entirety

**4.** Install the required libraries with the commands `go get github.com/go-sql-driver/mysql` and `go get github.com/bwmarrin/discordgo`.

**5.** Open your command prompt / terminal in the `/main` folder and type `go build`. You should see `main.exe` appear in the folder.

**6.** Create a file with the literal name of `token`. No extensions. There, you will put "Bot", then a space, then the token you get from [this](https://discordapp.com/developers/applications/me#top) page. Make sure to make your application a Bot User! A valid, but fake, token file would have a single line that looks like `Bot HdfgsdFGsGFD.34DFGsd33fE3FAKE`. Make sure this file is saved in `/main`.

**7.** Create a file called `db.auth` and put this inside: `root:<your password>@tcp(127.0.0.1:3306)/sweetiebot?parseTime=true&collation=utf8mb4_general_ci`. Save the file in `/main`.

**8.** Open the MySQL GUI under MariaDB called "HeidiSQL". Enter your password in the password box and then hit *Open*.

**9.** Top Bar: File > Run SQL File... then navigate to your cloned Sweetiebot folder and select "sweetiebot.sql". Afterwards, do the same with "sweetiebot_tz.sql".

**10.** Run `main.exe`.

You may have noticed that your bot isn't in your server yet. You will be able to connect your bot to your server by using this url.

`https://discordapp.com/oauth2/authorize?client_id=YOUR_BOT_CLIENT_ID&scope=bot`

Then add it to a server you have administration rights to.