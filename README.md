# qrn

qrn is a database load testing tool.

**NOTE: Currently developing the new tool. see https://github.com/winebarrel/qube**

## Usage

```
Usage of qrn:
  -commit-rate int
    	commit rate
  -data value
    	file path of execution queries for each agent
  -driver string
    	database driver
  -dsn string
    	data source name
  -force
    	ignore query error
  -hbins int
    	histogram bins (default 10)
  -hinterval string
    	histogram interval (default "0")
  -histogram
    	show histogram
  -html
    	output histogram html
  -key string
    	json key of query (default "query")
  -log string
    	file path of query log
  -logtime string
    	execution time threshold for logged queries (default "0")
  -loop
    	input data loop flag (default true)
  -maxcount int
    	maximum number of queries for each agent. zero is unlimited
  -nagents int
    	number of agents
  -pre-query value
    	queries to be pre-executed for each agent
  -query string
    	execution query
  -random value
    	randomize the start position of input data
  -rate int
    	rate limit for each agent (qps). zero is unlimited
  -time int
    	test run time (sec). zero is unlimited (default 60)
  -version
    	Print version and exit
```

```
$ echo '{"query":"select 1"}' >> data.jsonl
$ echo '{"query":"select 2"}' >> data.jsonl
$ echo '{"query":"select 3"}' >> data.jsonl
$ qrn -data data.jsonl -dsn root:@/ -nagents 4 -rate 5 -time 10 -histogram
00:07 | 4 agents / run 184 queries (20 qps)

          57µs - 115µs -
         115µs - 173µs -
         173µs - 231µs ---
         231µs - 289µs ------------
         289µs - 346µs ----------
         346µs - 404µs ------------------------------------------
         404µs - 462µs --------------------------------------------------------
         462µs - 520µs ------------------------------
         520µs - 760µs ----------

{
  "DSN": "root:@/",
  "Files": [
    "data.jsonl"
  ],
  "PreQueries": null,
  "Started": "2020-05-13T11:18:14.224848+09:00",
  "Finished": "2020-05-13T11:18:24.559912+09:00",
  "Elapsed": 10,
  "Queries": 189,
  "NAgents": 4,
  "Rate": 5,
  "QPS": 18.287694303306097,
  "MaxQPS": 21,
  "MinQPS": 18,
  "MedianQPS": 19,
  "ExpectedQPS": 20,
  "LoopCount": 15894,
  "Response": {
    "Time": {
      "Cumulative": "78.389862ms",
      "HMean": "392.47µs",
      "Avg": "414.761µs",
      "P50": "418.565µs",
      "P75": "462.099µs",
      "P95": "532.099µs",
      "P99": "735.68µs",
      "P999": "760.585µs",
      "Long5p": "632.823µs",
      "Short5p": "218.38µs",
      "Max": "760.585µs",
      "Min": "182.384µs",
      "Range": "578.201µs",
      "StdDev": "90.961µs"
    },
    "Rate": {
      "Second": 2411.0260584461803
    },
    "Samples": 189,
    "Count": 189,
    "Histogram": [
      {
        "57µs - 115µs": 1
      },
      {
        "115µs - 173µs": 1
      },
      {
        "173µs - 231µs": 4
      },
      {
        "231µs - 289µs": 14
      },
      {
        "289µs - 346µs": 12
      },
      {
        "346µs - 404µs": 48
      },
      {
        "404µs - 462µs": 63
      },
      {
        "462µs - 520µs": 34
      },
      {
        "520µs - 760µs": 12
      }
    ]
  },
  "Token": "a579889e-97f9-4fd1-8b33-93ab2c78e6ad",
  "GOMAXPROCS": 16
}
```

## DSN Examples

* https://github.com/go-sql-driver/mysql#examples
* https://github.com/jackc/pgx/blob/master/stdlib/sql.go

## Load different data for each agent

```
$ echo '{"query":"select 1"}' >> data1.jsonl
$ echo '{"query":"select 2"}' >> data2.jsonl
$ qrn -data data1.jsonl -data data2.json -dsn root:@/ -rate 5 -time 10 -histogram # -nagents 2
```

## Output Histogram HTML

If the `-html` is added, the histogram HTML will be output.

```
$ qrn -data data.jsonl -dsn root:@/ -nagents 8 -time 15 -html -hinterval 5ms -html
00:07 | 1 agents / run 654003 queries (78425 qps)
...

output qrn-1589336606.html
```

![](https://user-images.githubusercontent.com/117768/82013568-93bb6400-96b5-11ea-9001-cde7e2e50484.png)

## Related Links

* MySQL General Query Log parser
    * https://github.com/winebarrel/genlog
* qrn log analyzer
    * https://github.com/winebarrel/qrnlog
* MySQL load testing tools like mysqlslap that automatically generate test data
    * https://github.com/winebarrel/qlap
* Parser to extract SQL from postgresql.log
    * https://github.com/winebarrel/poslog
