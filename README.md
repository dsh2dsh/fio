# Fio banka report generator

This program reads a CSV file or stdin, processes it and outputs an aggreagated report of your expenses.

```
Usage:
  fio [input.csv] [flags]

Flags:
  -c, --config string      config file (default is .fio.yaml)
      --from-date string   skip payments before given date (in format YYYY-MM-DD)
  -h, --help               help for fio
  -m, --month string       include payments for given month (in format YYYY-MM)
      --to-date string     skip payments after given date (in format YYYY-MM-DD)
```

It expects a CSV file, downloaded from Fio banka, with next fields:

* Datum
* Objem
* Protiúčet
* Kód banky
* Zpráva pro příjemce
* Poznámka
* VS

Using a config file, `.fio.yaml` by default, it parses the CSV, aggregates transactions by rules from config file and outputs report to stdout.
