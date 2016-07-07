# coding: utf-8
require File.expand_path('../lib/ecfg/version', __FILE__)

files = File.read("MANIFEST").lines.map(&:chomp)

Gem::Specification.new do |spec|
  spec.name          = "ecfg"
  spec.version       = Ecfg::VERSION
  spec.authors       = ["Burke Libbey"]
  spec.email         = ["burke.libbey@shopify.com"]
  spec.summary       = %q{Asymmetric keywise encryption for configuration}
  spec.description   = %q{Secret management by encrypting values in a JSON or YAML file with a public/private keypair}
  spec.homepage      = "https://github.com/Shopify/ecfg"
  spec.license       = "MIT"

  spec.files         = files
  spec.executables   = ["ecfg"]
  spec.test_files    = []
  spec.require_paths = ["lib"]
end
