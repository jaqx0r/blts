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
