This reads the Influx database that is getting temperature inputs from each server using Telegraf

If any of the servers have a spike in temperature greater than n% (10 default)
it ends an alert to MQTT which is then picked up by Home Assistant and sent to my Atrix clock
