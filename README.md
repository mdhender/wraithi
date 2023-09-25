# Wraith

Wraith is a game engine and web server.

## Sources

Start the server with the ability to shut it down gracefully from
[clavinejune blog](https://clavinjune.dev/en/blogs/golang-http-server-graceful-shutdown/).

See
[Gregory Gaines' Blog](https://www.gregorygaines.com/blog/how-to-properly-hash-and-salt-passwords-in-golang-bcrypt/)
for details on why we want to use BCrypt.

See
[Scott Piper's Blog](http://0xdabbad00.com/2015/04/23/password_authentication_for_go_web_servers/)
for more details on auth/auth.

## Running as a system service

WARNING: Don't trust this application to be secure.
Use at your own risk.

I put together a `systemd` script based on the 
[DO Tutorial](https://www.digitalocean.com/community/tutorials/how-to-sandbox-processes-with-systemd-on-ubuntu-20-04).

    /etc/systemd/system# cat wraith.service
    [Unit]
    Description=Wraith server
    StartLimitIntervalSec=0
    After=network-online.target
    
    [Service]
    Type=simple
    User=www-data
    PIDFile=/run/wraith.pid
    WorkingDirectory=/var/www/wraith
    ExecStart=/usr/local/bin/wraith
    ExecReload=/bin/kill -USR1 $MAINPID
    Restart=on-failure
    RestartSec=1
    
    [Install]
    WantedBy=multi-user.target

Please do your own research on how to secure and lock down a service.

## Copying

This repository is licensed under the GNU Affero General Public License.
Please see the `LICENSE` file in the root of this repository.

