# NuSpider

NuSpider is a simple web crawler written in go. 

## Features
* Memory efficient
* Tracks visited urls
* Handles any sized site from personal blog to amazon
* Abides by robots.txt
* Accepts cookies
* Simple and easily customizable

## Limitations
* Crawls dont persist between runs (this can easily be added by persisting `current_id` and `first_id` in the kv store)
* Nothing in place that processes the crawled data

## Requirements
* Go
* [Glide](https://github.com/Masterminds/glide)

## Building
* glide install
* go build

## Usage
```
$ ./nuspider
Usage: ./nuspider <site>
```

Example:
```
$ ./nuspider www.nytimes.com
2017/05/23 20:54:12 Starting crawler for www.nytimes.com
2017/05/23 20:54:12 Fetching http://www.nytimes.com/
2017/05/23 20:54:13 Fetching http://www.nytimes.com/content/help/site/ie9-support.html
2017/05/23 20:54:13 Fetching http://www.nytimes.com/es/
2017/05/23 20:54:14 Fetching http://www.nytimes.com/pages/todayspaper/index.html
2017/05/23 20:54:15 Fetching http://www.nytimes.com/video
2017/05/23 20:54:15 Fetching https://www.nytimes.com/pages/world/index.html
2017/05/23 20:54:16 Fetching https://www.nytimes.com/pages/national/index.html
2017/05/23 20:54:16 Fetching https://www.nytimes.com/pages/politics/index.html
2017/05/23 20:54:17 Fetching https://www.nytimes.com/pages/nyregion/index.html
2017/05/23 20:54:18 Fetching https://www.nytimes.com/pages/business/index.html
2017/05/23 20:54:18 Fetching https://www.nytimes.com/pages/business/international/index.html
```
