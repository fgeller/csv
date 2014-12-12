csv -- cut for CSV files.
========================

N.B. I'm still playing with the main loop, things are doomed to fail ;)

Print selected comma separated values from each file to standard output.

csv is modeled after GNU coreutils' cut, but with CSV files in mind. The goal is
to process possibly large CSV files (>100MB) fairly efficiently and select
columns conveniently.

Examples
--------

    $ cat sample.csv
    first name,last name,favorite pet
    hans,hansen,moose
    peter,petersen,monarch
    $ go install github.com/fgeller/csv
    $ csv -c2-3 sample.csv
    last name,favorite pet
    hansen,moose
    petersen,monarch
    $ csv -n"favorite pet" sample.csv
    favorite pet
    moose
    monarch
    $ csv -n"favorite pet" --complement sample.csv
    first name,last name
    hans,hansen
    peter,petersen
    $ unix2dos sample.csv
    unix2dos: converting file sample.csv to DOS format...
    $ csv -N"last name" sample.csv 
    last name
    hansen
    petersen
    $ 


References
----------

* [cut invocation] (https://www.gnu.org/software/coreutils/manual/html_node/cut-invocation.html#cut-invocation)
* [RFC4180](https://www.rfc-editor.org/rfc/rfc4180.txt)
* [GNU coreutils](http://www.gnu.org/software/coreutils/)
