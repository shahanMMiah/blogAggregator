
<h1 align="center">
  <br>
  BlogAggregator
  <br>
</h1>

<h4 align="center">A CLI tool for finding/viewing blogpost's from followed rssfeeds</a>.</h4>

<p align="center">

<p align="center">
  <a href="#key-features">Key Features</a> •
  <a href="#how-to-use">How To Use</a> •
  <a href="#credits">Credits</a> •
  <a href="#license">License</a>
</p>

## Key Features

* Registering and Logging users
* Saving/fetching feeds and post data to and from a database   
* Following and Unfollowing rss feeds 
* Web Sraping RSS feeds for new posts every specifeid time intevral



## How To Use

To Run this application, you'll need to install 
[Golang](https://go.dev/doc/install) 
[Postgresql](https://www.postgresql.org/docs) 
[Goose](https://github.com/pressly/goose/)



Set up the gator database
```bash
sudo apt install postgresql postgresql-contrib
sudo service postgresql start
sudo -u postgres psql
CREATE DATABASE gator;
```

database migrations and tool install
```bash
$ go install github.com/pressly/goose/v3/cmd/goose@latest
$ git clone https://github.com/shahanMMiah/blogAggregator.git
$ cd blogAggregator/sql/schema
$ goose postgres postgres://postgres:postgres@localhost:5432/gator up
$ go install
```

Setting up a config file: Create a json file in home dir name '.gatorconfig.json" with contents:
```json
'{"Db_url":"postgres://postgres:postgres@localhost:5432/gator?sslmode=disable","Current_user_name":"","Posts_limit":10}'
```

```bash
# tool usage exmaples:
$ blogAggregator help # veiw all commands help
$ blogAggregator register {username} # register username
$ blogAggreator addfeed {rss xml URL} # add/follow rss feed 
$ blogAggregator agg 4s #scape feeds for posts ever 4 seconds
$ blogAggregator browse 10 #view 10 of the saved posts from scaped feeds 
```


## Credits

This software uses the following open source packages, postgresql, goose:

- [Golang](https://go.dev)
- [Postgresql](https://www.postgresql.org) 
- [Goose](https://github.com/pressly/goose/)


## License

MIT
---

> GitHub [@shahanMMiah](https://github.com/shahanMMiah)

