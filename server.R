#!/usr/bin/env Rscript
#
# Copyright 2018 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

library(zoo)

u <- function(x) as.POSIXct(x, origin="1970-01-01", tz="UTC")
r <- function(line) read.zoo(textConnection(line), sep=",", header=FALSE, FUN=u)

# waits for connect
sock <- make.socket("localhost", 10000, server = TRUE)
on.exit(close.socket(sock))


lmz <- r(read.socket(sock))

repeat {
  line <- read.socket(sock, loop = TRUE)
  if (line == "") break

  data <- r(line)

  lmz <- rbind(lmz, data)
}

cat(lmz)
