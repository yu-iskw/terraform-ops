class TerraformOps < Formula
  desc "A command-line interface for managing Terraform operations"
  homepage "https://github.com/yu/terraform-ops"
  version "0.1.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yu/terraform-ops/releases/download/v0.1.0/terraform-ops-darwin-arm64"
      sha256 "PLACEHOLDER_SHA256_ARM64"
    else
      url "https://github.com/yu/terraform-ops/releases/download/v0.1.0/terraform-ops-darwin-amd64"
      sha256 "PLACEHOLDER_SHA256_AMD64"
    end
  end

  on_linux do
    url "https://github.com/yu/terraform-ops/releases/download/v0.1.0/terraform-ops-linux-amd64"
    sha256 "PLACEHOLDER_SHA256_LINUX"
  end

  def install
    bin.install Dir["terraform-ops-*"].first => "terraform-ops"
  end

  test do
    system "#{bin}/terraform-ops", "--help"
  end
end
