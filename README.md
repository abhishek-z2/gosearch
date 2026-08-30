# gosearch

A command-line text search utility written in Go.  

`gosearch` searches files for a given text pattern and supports recursive  
directory searching, case-insensitive matching, inverted matching, context  
lines, extension filtering, match counting, and colored output.  
- ## Features
- Search individual files
- Recursively search directories
- Case-insensitive matching with `-i`
- Invert matches with `-v`
- Count matches with `-c`
- Filter files by extension with `-e`
- Show lines before a match with `-B`
- Show lines after a match with `-A`
- Highlight matches with `-color`
- Concurrent directory searching using a worker pool
- Skips `.git` directories
- Skips binary files
- ## Installation
- ### Build from source

Make sure you have Go installed, then clone the repository:  

```bash
git clone <your-repository-url>
cd gosearch
go build -o gosearch
```

You can then run:  

```
./gosearch <query> <filepath>
```
- ## Usage
- ### Search a file

```
./gosearch "hello" file.txt
```
- ### Search a directory recursively

```
./gosearch -r "hello" ./project
```
- ### Case-insensitive search

```
./gosearch -i "hello" file.txt
```
- ### Invert the match

```
./gosearch -v "hello" file.txt
```

This prints lines that do **not** contain the query.  
- ### Count matches

```
./gosearch -c "hello" file.txt
```
- ### Search specific file extensions

```
./gosearch -r -e go "SearchFile" ./project
```

Multiple extensions can be provided as a comma-separated list:  

```
./gosearch -r -e go,txt "hello" ./project
```
- ### Show context around matches

Show 2 lines before each match:  

```
./gosearch -B 2 "hello" file.txt
```

Show 2 lines after each match:  

```
./gosearch -A 2 "hello" file.txt
```

Both can be combined:  

```
./gosearch -B 2 -A 2 "hello" file.txt
```
- ### Colored output

```
./gosearch -color "hello" file.txt
```
- ## Command-line options

| Option | Description |
|---|---|
| `-r` | Search directories recursively |
| `-i` | Perform a case-insensitive search |
| `-v` | Select non-matching lines |
| `-c` | Return only the number of matches |
| `-e` | Search only files with the specified extension |
| `-A N` | Print `N` lines after a match |
| `-B N` | Print `N` lines before a match |
| `-color` | Enable colored output |
- ## How it works

For a single file, `gosearch` reads the file line by line using a buffered  
scanner and checks each line for the requested query.  

Directory searches use a concurrent worker-pool design.  

```
              filepath.WalkDir                     
                     │
                     │ file paths
                     ▼
                ┌───────────┐
                │   jobs    │
                │  channel  │
                └─────┬─────┘
                      │
         ┌────────────┼────────────┐
         ▼            ▼            ▼
    ┌────────┐   ┌────────┐   ┌────────┐
    │ Worker │   │ Worker │   │ Worker │
    │   1    │   │   2    │   │   N    │
    └───┬────┘   └───┬────┘   └───┬────┘
        │            │            │
        └────────────┼────────────┘
                     ▼
                ┌───────────┐
                │  results  │
                │  channel  │
                └─────┬─────┘
                      │
                      ▼
                PrintResult()

```
The directory walker acts as the producer, placing searchable file paths  
onto the `jobs` channel.  

A fixed number of worker goroutines consume those paths and perform the  
search using `SearchFile`.  

The workers send their results through the `results` channel, which is  
consumed by the main search routine and passed to `PrintResult`.  

A `sync.WaitGroup` is used to determine when all workers have completed,  
  after which the results channel is closed.  
  - ## Benchmark

  The project includes benchmarks comparing sequential and concurrent  
  directory searching.  

  The benchmark workload searches a directory containing 500 files.  

  Example benchmark results:  

  ```
  BenchmarkSearchDirectory-4
  ~10.62 ms/op
  ~33.14 MB/op
  ~8,559 allocs/op

  BenchmarkSearchDirectoryConcurrent-4
  ~8.12 ms/op
  ~33.17 MB/op
  ~8,590 allocs/op
  ```

  On this benchmark workload, the concurrent implementation was approximately  
  23% faster than the sequential implementation.  

  The benchmark was run on:  

  ```
  OS:         Linux
  Architecture: amd64
  CPU:        Intel Core i3-1005G1 @ 1.20GHz
  ```

  These results are workload- and machine-dependent and should not be  
  interpreted as a general performance guarantee.  
  - ## Testing

  Run the test suite with:  

  ```
  go test ./...
  ```

  Run benchmarks with:  

  ```
  go test -bench=. -benchmem ./search
  ```
  - ## Project structure

  ```
  gosearch/
  ├── main.go
  ├── main_test.go
  ├── go.mod
  └── search/
      ├── options.go
      ├── result.go
      ├── search.go
      └── search_test.go
  ```
   


