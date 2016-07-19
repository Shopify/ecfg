# ecfg

`ecfg` is a utility for managing a collection of secrets, generally with the
intent of committing them to source control. The secrets are encrypted using
[public key](http://en.wikipedia.org/wiki/Public-key_cryptography), [elliptic
curve](http://en.wikipedia.org/wiki/Elliptic_curve_cryptography) cryptography
([NaCl](http://nacl.cr.yp.to/) [Box](http://nacl.cr.yp.to/box.html):
[Curve25519](http://en.wikipedia.org/wiki/Curve25519) +
[Salsa20](http://en.wikipedia.org/wiki/Salsa20) +
[Poly1305-AES](http://en.wikipedia.org/wiki/Poly1305-AES)). Secrets are
collected in a JSON, YAML, or TOML file, in which all the string values are encrypted.
Public keys are embedded in the file, and the decrypter looks up the
corresponding private key from its local filesystem or the process environment.

The main benefits provided by `ecfg` are:

* Secrets can be safely stored in a git repo.
* Changes to secrets are auditable on a line-by-line basis with `git blame`.
* Anyone with git commit access has access to write new secrets.
* Decryption access can easily be locked down to production servers only.
* Secrets change synchronously with application source (as opposed to secrets
  provisioned by Configuration Management).
* Simple, well-tested, easily-auditable source.

See [the manpages](https://shopify.github.io/ecfg) for more technical documentation.

*ecfg is written in lowercase, and normally pronounced "e-config", like "email".*

## Differences from EJSON 1.0

* Supports YAML and TOML
* Expected filename changes from `*.ejson` to `*.ecfg.<format>` (e.g.
  `.ecfg.json`)
* `encrypt` and `decrypt` can now both receive data on `stdin`, and will emit
  output to `stdout` in this case.
* Added `--type`/`-t` flag to `encrypt` and `decrypt` commands. If the filename
  is `*{json,yaml,yml,toml}`, this flag is optional, but when the type can't
  be inferred because of a different filename or when reading from stdin, it
  must be specified. (note that this recognizes `*.ejson` as json automatically)
* All references to EJSON have changed to ECFG, including `ECFG_KEYDIR`.
* Key lookup has been changed to allow a greater degree of control:
* `ECFG_PRIVATE_KEY` may be set as an environment variable. This can be used to
  preempt key selection logic during decrypt. If present, ecfg will blindly
  attempt to decrypt any encrypted values using this key, instead of trying to
  find the matching key in the keydir. This is useful for deploying to Heroku.
* Unless `ECFG_KEYDIR` is set, ecfg will now be look in four directories (in
  order of decreasing precedence):
  * `$XDG_CONFIG_HOME/ecfg/keys` (if set, and non-root)
  * `$HOME/.ecfg/keys` (if non-root)
  * `/etc/ecfg/keys`
  * `/opt/ejson/keys`
* If `ECFG_KEYDIR` is set, then ecfg will look in `$ECFG_KEYDIR/keys`
* Overhauled build process

## Installation

You can download the `.deb` package from [Github
Releases](https://github.com/Shopify/ecfg/releases).

On macOS, you can install ecfg using `brew install shopify/shopify/ecfg`.

You can also install ecfg as a gem using `gem install ecfg` or adding it to
your `Gemfile`.

## Workflow

### 1: Create the Keydir

By default, ecfg looks for keys in `/etc/ecfg/keys`. You can change this by
setting `ECFG_KEYDIR` or passing the `--keydir` option.

```
$ mkdir -p /etc/ecfg/keys
```

### 2: Generate a keypair

When called with `-w`, `ecfg keygen` will write the keypair into the `keydir`
and print the public key. Without `-w`, it will print both keys to stdout. This
is useful if you have to distribute the key to multiple servers via
configuration management, etc.

`ecfg keygen` will write the keypair to the first directory on the keypath
that's writable:

1. `$XDG_CONFIG_HOME/ecfg/keys` (if non-root and `$XDG_CONFIG_HOME` is set)
1. `$HOME/.ecfg/keys` (if non-root)
1. `/etc/ecfg/keys`
1. `/opt/ejson/keys`

```
$ ecfg keygen
Public Key:
63ccf05a9492e68e12eeb1c705888aebdcc0080af7e594fc402beb24cce9d14f
Private Key:
75b80b4a693156eb435f4ed2fe397e583f461f09fd99ec2bd1bdef0a56cf6e64
```

```
$ ecfg keygen -w
53393332c6c7c474af603c078f5696c8fe16677a09a711bba299a6c1c1676a59
$ cat ~/.ecfg/keys/5339*
888a4291bef9135729357b8c70e5a62b0bbe104a679d829cdbe56d46a4481aaf
```

```
$ sudo ecfg keygen -w
53393332c6c7c474af603c078f5696c8fe16677a09a711bba299a6c1c1676a59
$ cat /etc/ecfg/keys/5339*
888a4291bef9135729357b8c70e5a62b0bbe104a679d829cdbe56d46a4481aaf
```

### 3: Create an `ecfg` file

The format is described in more detail [later on](#format). For now, create a
file that looks something like this. Fill in the `<key>` with whatever you got
back in step 2.

Create this file as `test.ecfg.json`:

```json
{
  "_public_key": "<key>",
  "database_password": "1234password"
}
```

### 4: Encrypt the file

Running `ecfg encrypt test.ecfg.json` will encrypt any new plaintext keys in the
file, and leave any existing encrypted keys untouched:

```json
{
  "_public_key": "63ccf05a9492e68e12eeb1c705888aebdcc0080af7e594fc402beb24cce9d14f",
  "database_password": "EJ[1:WGj2t4znULHT1IRveMEdvvNXqZzNBNMsJ5iZVy6Dvxs=:kA6ekF8ViYR5ZLeSmMXWsdLfWr7wn9qS:fcHQtdt6nqcNOXa97/M278RX6w==]"
}
```

Try adding another plaintext secret to the file and run `ecfg encrypt
test.ecfg.json` again. The `database_password` field will not be changed, but
the new secret will be encrypted.

### 5: Decrypt the file

To decrypt the file, you must have a file present in the `keydir` whose name is
the 64-byte hex-encoded public key exactly as embedded in the `ecfg` document.
The contents of that file must be the similarly-encoded private key. If you
used `ecfg keygen -w`, you've already got this covered.

Alternatively, in some environments, it may be easier to pass the private key
via `ECFG_PRIVATE_KEY`, which preempts the `keydir` lookup.

Unlike `ecfg encrypt`, which overwrites the specified files, `ecfg decrypt`
only takes one file parameter, and prints the output to `stdout`:

```
$ ecfg decrypt foo.ecfg.json
{
  "_public_key": "63ccf05a9492e68e12eeb1c705888aebdcc0080af7e594fc402beb24cce9d14f",
  "database_password": "1234password"
}
```

## Format

The `ecfg.json` document format is simple, but there are a few points to be aware
of:

1. It's just JSON (or YAML or TOML, in the case of `ecfg.yaml` and `ecfg.toml`)
2. There *must* be a key at the top level named `_public_key`, whose value is a
   32-byte hex-encoded (i.e. 64 ASCII byte) public key as generated by `ecfg
   keygen`.
3. Any string literal that isn't an object key will be encrypted by default (ie.
   in `{"a": "b"}`, `"b"` will be encrypted, but `"a"` will not.
4. Numbers, booleans, and nulls aren't encrypted.
5. If a key begins with an underscore, its corresponding value will not be
   encrypted. This is used to prevent the `_public_key` field from being
   encrypted, and is useful for implementing metadata schemes.
6. Underscores do not propagate downward. For example, in `{"_a": {"b": "c"}}`,
   `"c"` will be encrypted.

## Building ecfg

**If you work at Shopify, just run `dev up && dev build`; otherwise:**

1. Install ruby *(the system one will do on OS X, or you can `brew install ruby`. `hpricot` doesn't seem to want to build with 2.3.x)*
1. Install bundler *(`gem install bundler`)*
1. `bundle install`
1. Install Go *(`brew install go`)*
1. [Configure your `$GOPATH`](https://github.com/golang/go/wiki/GOPATH) and make sure this is on it at github.com/Shopify/ecfg.
1. `make`
