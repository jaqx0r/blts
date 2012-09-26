#!/usr/bin/env Rscript

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
