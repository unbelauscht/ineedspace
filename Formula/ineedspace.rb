class Ineedspace < Formula
  desc "Interactive TUI and CLI disk cleaner for developers and sysadmins"
  homepage "https://github.com/unbelauscht/ineedspace"
  url "https://github.com/unbelauscht/ineedspace/releases/download/v0.2.1/ineedspace-darwin-universal.tar.gz"
  sha256 "bbd90e59c3ff956c78c39d9012763f9c9420f7d377fe0b2601673e6ccd566464"
  license "MIT"

  def install
    bin.install "ineedspace-darwin-universal" => "ineedspace"
  end

  test do
    assert_match "ineedspace", shell_output("#{bin}/ineedspace --version")
  end
end
