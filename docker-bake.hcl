target "pijar-app" {
  context = "."
  dockerfile = "Dockerfile"
  tags = ["pijar-app:latest"]
}

target "pijar-db" {
  context = "./"
  dockerfile = "Dockerfile"
  tags = ["pijar-db:latest"]
}