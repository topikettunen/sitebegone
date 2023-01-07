# sitebegone

Block distracting sites for good.

I have a terrible habit of losing endless hours to distracting
websites. This is a habit that I tend to be conscious of. I have noticed
that extreme measures tend to work best when I'm trying to work on
something, so I wrote this simple application to completely block these
kinds of sites. This doesn't block them entirely since it only adds a
new entry to your `hosts` file, so when you go to the site to be
blocked, it'll forward you to `localhost`. Still, it adds an extra
measure to prevent you from wasting hours on these sites. 

After you run, for example, `sitebegone youtube.com`, this application
adds the following section to your `hosts` file: 

```alpha
# Added by sitebegone

127.0.0.1	youtube.com
127.0.0.1	www.youtube.com

# End of sitebegone section
```

Successive sites to be blocked would then be added inside this block so
the application knows what entries it should handle. There is no removal
of these entries, so you need to edit the `hosts` file manually to
remove the blocking. 

## Install

```shell
$ go get github.com/topikettunen/sitebegone
```

## Usage

```shell
$ sudo sitebegone youtube.com
```
