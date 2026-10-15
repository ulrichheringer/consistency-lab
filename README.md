# consistency-lab

A browser demo of three replicas that receive writes after a delay. Write different values to the same key to watch the replicas diverge, report a conflict, and converge on the latest write.

Install with go install github.com/ulrichheringer/consistency-lab@latest.
Run consistency-lab and open http://localhost:8090.
