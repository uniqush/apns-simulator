Device Token Generator
==========

This little program can randomly generate device token, which can be in turn used
for the APNS simulator.

Example:

        $ ./devtokengen -n 10
        curl http://127.0.0.1:9898/subscribe -d service=simserv -d subscriber=usr.0 -d pushservicetype=apns -d devtoken=0000000000000000000000000000000000000000000000004d65822107fcfd53
        curl http://127.0.0.1:9898/subscribe -d service=simserv -d subscriber=usr.1 -d pushservicetype=apns -d devtoken=0000000000000000000000000000000000000000000000004d65822107fcfd54
        curl http://127.0.0.1:9898/subscribe -d service=simserv -d subscriber=usr.2 -d pushservicetype=apns -d devtoken=0000000000000000000000000000000000000000000000004d65822107fcfd55
        curl http://127.0.0.1:9898/subscribe -d service=simserv -d subscriber=usr.3 -d pushservicetype=apns -d devtoken=0000000000000000000000000000000000000000000000004d65822107fcfd56
        curl http://127.0.0.1:9898/subscribe -d service=simserv -d subscriber=usr.4 -d pushservicetype=apns -d devtoken=0000000000000000000000000000000000000000000000004d65822107fcfd57
        curl http://127.0.0.1:9898/subscribe -d service=simserv -d subscriber=usr.5 -d pushservicetype=apns -d devtoken=0000000000000000000000000000000000000000000000004d65822107fcfd58
        curl http://127.0.0.1:9898/subscribe -d service=simserv -d subscriber=usr.6 -d pushservicetype=apns -d devtoken=0000000000000000000000000000000000000000000000004d65822107fcfd59
        curl http://127.0.0.1:9898/subscribe -d service=simserv -d subscriber=usr.7 -d pushservicetype=apns -d devtoken=0000000000000000000000000000000000000000000000004d65822107fcfd5a
        curl http://127.0.0.1:9898/subscribe -d service=simserv -d subscriber=usr.8 -d pushservicetype=apns -d devtoken=0000000000000000000000000000000000000000000000004d65822107fcfd5b
        curl http://127.0.0.1:9898/subscribe -d service=simserv -d subscriber=usr.9 -d pushservicetype=apns -d devtoken=0000000000000000000000000000000000000000000000004d65822107fcfd5c

The output can be directly used with [uniqush-push]. If you are running [uniqush-push]:

        $ ./devtokengen -n 10 | sh

[uniqush-push]: http://github.com/uniqush/uniqush-push

