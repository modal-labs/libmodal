require 'modal'

def main
  fn = Modal::Function_.lookup("my-modal-app", "echo_string")
  result = fn.remote([], {message: "Hello, Modal!"})
  puts "Function call result: #{result}"
end

main
