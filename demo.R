library(zoo)

u <- function(x) as.POSIXct(x, origin="1970-01-01", tz="UTC")
r <- function(x, y) read.zoo(paste("data/:", x, "-", y, ".csv", sep=""), sep=",", header=FALSE, FUN=u)

library(mail)

alert <- function(message, detail) sendmail('jaq@spacepants.org',subject=message, message=detail)

ez <- r("9001", "errors")
ez
diff(ez)

rz <- r("9001", "requests")
diff(rz)

plot(diff(rz))
plot(diff(ez)/diff(rz))

er <- diff(ez)/diff(rz)
er

er[er > 0.2]

if (er[er > 0.2]) { er[er > 0.2] }

if (er[er > 0.2]) { alert("error rate high") }


lmz <- r("9001", "latency_ms")
plot(diff(lmz), plot.type="single", col=c(1:10))

plot(rowSums(diff(lmz)[,1:9])/rowSums(diff(lmz)))


over256 <- rowSums(diff(lmz[,8:10]))/rowSums(diff(lmz))
