library(zoo)

u <- function(x) as.POSIXct(x, origin="1970-01-01", tz="UTC")
r <- function(x, y) read.zoo(paste("data/:", x, "-", y, ".csv", sep=""), sep=",", header=FALSE, FUN=u)

library(mail)

alert <- function(message, detail) sendmail('YOU at example.com',subject=message, message=detail)

# ---

# Load errors
ez <- r("9001", "errors")
ez

diff(ez)

# load requests
rz <- r("9001", "requests")
diff(rz)

plot(diff(rz))

# error ratio
er <- diff(ez)/diff(rz)
er <- er[is.finite(er)]

plot(er)

# filter high error ratio
er[er > 0.2]

if (length(er[er > 0.2] > 0)) { er[er > 0.2] }

# alert!
if (length(er[er > 0.2] > 0)) { alert("error rate high", er[er > 0.2]) }


# latency plots
lmz <- r("9001", "latency_ms")

plot(diff(lmz), plot.type="single", col=c(1:10))

plot(rowSums(diff(lmz)[,1:9])/rowSums(diff(lmz)))

over256 <- rowSums(diff(lmz[,8:10]))/rowSums(diff(lmz))

plot(over256)
