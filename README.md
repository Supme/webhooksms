# webhooksms
Webhook for send SMS from SMPP transport

Support Grafana, Prometheus webhook send SMS

## Install

Download latest version from https://github.com/Supme/webhooksms/releases
```
tar -xvzf webhooksms-v0.x.x.tar.gz
useradd --no-create-home --shell /bin/false webhooksms
cp webhooksms /usr/local/bin/
chown webhooksms: /usr/local/bin/webhooksms
mkdir /var/log/webhooksms
chown webhooksms: /var/log/webhooksms
mkdir /etc/webhooksms
cp config.ini.example /etc/webhooksms/config.ini
chown -R webhooksms: /etc/webhooksms
cp webhooksms.service /etc/systemd/system/
systemctl daemon-reload
```
edit /etc/webhooksms/config.ini
```
systemctl enable webhooksms
systemctl start webhooksms
```

## Use
Send webhook method POST and Basic Auth (user and password from config):
- Grafana (tested)
```
http://host/grafana
```
- Prometheus (not tested)
```
http://host/prometheus
```
