require 'rake'

Gem::Specification.new do |s|
  s.name = "modal"
  s.version = "0.0.0"
  s.required_ruby_version = ">= 2.7.0"
  s.summary = "Interact with modal from your Ruby code"
  s.description = "A gem to interact with Modal from your Ruby, Rails, or Sinatra applications"
  s.authors = ["Anthony Corletti"]
  s.email = ["anthcor@gmail.com"]
  s.files = FileList["lib/modal.rb", "lib/modal/*.rb", "lib/modal_proto/*.rb"].to_a
  s.homepage = "https://rubygems.org/gems/modal_rb"
  s.license = "Apache-2.0"
end
