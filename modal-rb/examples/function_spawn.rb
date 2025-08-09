require 'modal'

def main
  fn = Modal::Function_.lookup("my-modal-app", "echo_string")
  result = fn.spawn([], {message: "Hello, Modal!"})
  puts "Function spawn result: #{result}"
end

main
