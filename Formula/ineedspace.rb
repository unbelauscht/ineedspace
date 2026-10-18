class Ineedspace < Formula
  desc "Interactive TUI and CLI disk cleaner for developers and sysadmins"
  homepage "https://github.com/unbelauscht/ineedspace"
  version "0.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/unbelauscht/ineedspace/releases/download/v#{version}/ineedspace-darwin-arm64.tar.gz"
      sha256 "CHECKSUM_ARM64_PLACEHOLDER"

      def install
        bin.install "ineedspace-darwin-arm64" => "ineedspace"
      end
    else
      url "https://github.com/unbelauscht/ineedspace/releases/download/v#{version}/ineedspace-darwin-amd64.tar.gz"
      sha256 "CHECKSUM_AMD64_PLACEHOLDER"

      def install
        bin.install "ineedspace-darwin-amd64" => "ineedspace"
      end
    end
  end

  test do
    assert_match "ineedspace", shell_output("#{bin}/ineedspace --version")
  end
end
