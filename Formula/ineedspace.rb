class Ineedspace < Formula
  desc "Interactive TUI and CLI disk cleaner for developers and sysadmins"
  homepage "https://github.com/unbelauscht/ineedspace"
  version "0.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/unbelauscht/ineedspace/releases/download/v#{version}/ineedspace-darwin-arm64.tar.gz"
      sha256 "8e5309baa805b06827afeafed649016a43b7a95449c51a581fd80e9161986557"

      def install
        bin.install "ineedspace-darwin-arm64" => "ineedspace"
      end
    else
      url "https://github.com/unbelauscht/ineedspace/releases/download/v#{version}/ineedspace-darwin-amd64.tar.gz"
      sha256 "a044435a6c0471ee5a0de1beb38a20c2348502ba736ebfbd5ad4c03ed3725a43"

      def install
        bin.install "ineedspace-darwin-amd64" => "ineedspace"
      end
    end
  end

  test do
    assert_match "ineedspace", shell_output("#{bin}/ineedspace --version")
  end
end
