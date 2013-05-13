# What is this?

Chronoboost is a little web app that reads official
live-timing streams from formula1.com and displays
them in a web UI. It's an inferior port of Scott James
Remnant's excellent live-f1 (https://launchpad.net/live-f1)
to Go and the Web.


# Installation

Assuming you have installed the Go SDK, you should be
able to run something like:

  $ go install github.com/jobi/chronoboost/chronoboost
  $ cd ${GOPATH}/github.com/jobi/chronoboost/chronoboost
  $ chronoboost -email <your formula1.com account> -password <your formula1.com password>

Then open a browser at http://localhost:8080/
