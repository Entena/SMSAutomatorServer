# What is this?
This is the "worker" of my project. A good old fashioned dumb android app that constantly queries a known url looking for a json payload.
If it gets that payload then it triggers conditions to send an SMS on the device's behalf. Think of this as something you'd hook up in your
home business or lab that you want to send messages to.

# Novel use cases I've thought of
* I'm a small business and want fast and easy ways to send quick sms updates to clients
* I'm a security concious user hosting my own email services. I'd like to keep my 2FA in house so I use this to have my own node

## Anything To Know
Current design implements a start variable timestamp. The design also isn't great at catchign seen messages and will requeue. A simple fix later
should solve that. The UI is simple on purpose, you have a start button, you have a box showing the last response you got from the API.
