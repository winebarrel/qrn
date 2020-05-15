# qrn

qrn is a database benchmarking tool.
Currently only MySQL is supported.

## Usage

```
Usage of ./qrn:
  -data string
    	file path of execution queries
  -dsn string
    	data source name
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
  -loop
    	input data loop flag (default true)
  -nagents int
    	number of agents (default 1)
  -random
    	randomize the start position of input data (default true)
  -rate int
    	rate limit for each agent (qps). zero is unlimited
  -script string
    	file path of execution script
  -time int
    	test run time (sec) (default 60)
  -version
    	Print version and exit
```

```
$ echo '{"query":"select 1"}' >> data.jsonl
$ echo '{"query":"select 2"}' >> data.jsonl
$ echo '{"query":"select 3"}' >> data.jsonl
$ qrn -file data.jsonl -dsn root:@/ -nagents 4 -rate 5 -time 10 -histogram
/ run 184 queries (20 qps)

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
  "Started": "2020-05-13T11:18:14.224848+09:00",
  "Finished": "2020-05-13T11:18:24.559912+09:00",
  "Elapsed": 10,
  "Queries": 189,
  "NAgent": 4,
  "Rate": 5,
  "QPS": 18.287694303306097,
  "ExpectedQPS": 20,
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
  }
}
```

## DSN Examples

see https://github.com/go-sql-driver/mysql#examples

## Use Script as Data

```
$ cat data.js
for (var i = 0; i < 10; i++) {
  query("select " + i);
}

$ qrn  -script data.js -dsn root:@/ -nagents 8 -time 15 -rate 5
```

## Output histogram HTML

If the `-html` is added, the histogram HTML will be output.

```
$ qrn -data data.jsonl -dsn root:@/ -nagents 8 -time 15 -html -hinterval 1ms -html
- run 654003 queries (78425 qps)
...

output qrn-1589336606.html
```

![](https://user-images.githubusercontent.com/117768/81766121-78632400-9510-11ea-898a-83248aa5faeb.png)
