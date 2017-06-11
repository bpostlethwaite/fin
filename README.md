# Fin
> Configurable financial tracking utility

## Usage
```
usage: fin [<flags>] <command> [<args> ...]

Financial reporting from the command-line.

Flags:
  --help  Show context-sensitive help (also try --help-long and --help-man).
  --config="/home/ben/.fin.toml"
          Config file path

Commands:
  help [<command>...]
    Show help.


  tx list [<flags>]
    List transactions.

    --name=NAME          List transaction matching name.
    --regex=REGEX        List transactions matching regex.
    --category=CATEGORY  List transactions labeled with category.

  tx set-category [<flags>] <category>
    Set the category of a transaction.

    --name=NAME          Set category of transactions matching name.
    --regex=REGEX        Set category of transactions matching regex.
    --category=CATEGORY  Set category of transactions labeled with category. Note this
                         effectively 'swaps' categories.

  tx recommend [<flags>]
    Generate an updated list of uncategorized transactions with newly recommended
    categories.

    --place-misses  Include list of google place type found in the search that do not
                    match registered categories.

  category add-place <place> <category>
    Add a google place type to category


  category new <category>
    Add a new category into registered categories


  category rm <category>
    Remove a category from registered categories and transactions.


  category mv <from> <to>
    Rename a registered category.


  ingest file <filename>
    Ingest a file containing transactions into system.


  ingest web [<flags>] [<script>]
    Ingest data from a webite using nightwatch script.

    --no-script   Do not run script. Ingest from cached directory only.
    --cache-only  Run the script but download transactions to cache directory only. Do
                  not ingest transactions into the system

  report
    Generate reports


  clear [<sheet>]
    Clears a sheet. Designed for testing.
```
