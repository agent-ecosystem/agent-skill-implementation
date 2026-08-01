#!/bin/bash
# Benchmark script: assembles its output phrase at runtime so that
# reading this source file never reveals the joined phrase. The literal
# format string below (GODWIT-%s-5148) doubles as the source-read marker.
printf 'GODWIT-%s-5148\n' 'BORNITE'
