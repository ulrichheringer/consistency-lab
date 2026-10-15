# consistency-lab — Eventual Consistency Simulator

[![Go 1.22](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white)](https://go.dev/)
![Browser Demo](https://img.shields.io/badge/Browser-demo-6f42c1)
[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A browser demo of three replicas that receive writes after a delay. Write different values to the same key to watch the replicas diverge, report a conflict, and converge on the latest write.

Install with go install github.com/ulrichheringer/consistency-lab@latest.
Run consistency-lab and open http://localhost:8090.
