class Ineedspace < Formula
  desc "Interactive TUI and CLI disk cleaner for developers and sysadmins"
  homepage "https://github.com/unbelauscht/ineedspace"
  url "https://github.com/unbelauscht/ineedspace/releases/download/v0.1.0/ineedspace-darwin-universal.tar.gz"
  sha256 "b0ec459652d9fab5d8bb47432787837576a01918237401a6f2ef6917cf5f67ff"
  version "0.1.0"
  license "MIT"

  def install
    bin.install "ineedspace-darwin-universal" => "ineedspace"
  end

  test do
    assert_match "ineedspace", shell_output("#{bin}/ineedspace --version")
  end
end
