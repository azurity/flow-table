# flow-table

flow-table is a template rendering application / library for `.xlsx` file. Pure golang implementation.

## HOW TO USE

```
./flow-table -template <template xlsx file> -data <folder containing data> -output <output file>
```

## template grammar

Write in any table cell:
```
{{H(.3f)|[js] [data.x, data.y] }}
```

Next, flow-table will use the `js` language to execute the expression `[data.x, data.y]` and render first element in this cell, and extended horizontally (render the second element in right cell), inheriting the same style, and using `.3f` as the number format.

more abbr:
```
{{C|[js] data.x }}
{{[js] data.x }}
```

- expand direction:
    - `C`: only render in this Cell
    - `H`: fill cells Horizontally
    - `V`: fill cells Vertically
    - `T`: fill cells all direct
- format:
    - `s`: as a string
    - `Xd`: such as `10d`, integer
    - `.Xf`: such as `.3f`, float with 3 digits after the decimal point character
    - `.Xp`: such as `.2p`, percentage with 2 digits after the decimal point character
- language:
    allow some alias name, `javascript` can also be use as `js`

## data input type

flow-table will try load all files in the given directory for using as data source (Won't search file recurse). Each loaded data source will use the file name (without the extension) as the variable name.

support format:
- sqlite:
    - with ext `.db` or `.sqlite`
    - each table will be a property(in javascript) or key-value(in other language)
    - each property is an array of struct / dict
- xlsx:
    - with ext `.xlsx`
    - each table will be a property
    - each property is a 2d-array
- csv:
    - value is a 2d-array

## support script language

- javascript:
    - impl based on [goja](https://github.com/dop251/goja)
- python:
    - impl based on [gpython](https://github.com/go-python/gpython)
- cel:
    - impl based on [cel-go](https://github.com/google/cel-go)

> If you want more language support, please create an Issue or PR.
