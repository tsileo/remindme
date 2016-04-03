# remindme

Send simple reminder to your Pebble using the command line

## Getting Started

Download an app like [Timeline Token](http://apps.getpebble.com/en_US/application/5648acf2b2013fe638000097) to retrieve your own token.

```shell
$ export PEBBLE_TOKEN=yourtoken
$ remindme 2016-04-21T21:19 my things to remember
# vim will spwan and you will be able to write the reminder body here
#
# you also specify a delay
$ remindme 3h do this thing
```
