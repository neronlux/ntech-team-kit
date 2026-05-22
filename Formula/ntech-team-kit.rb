class NtechTeamKit < Formula
  desc "OpenCode-native skills, agents, commands, and rules (Cursor Team Kit port)"
  homepage "https://github.com/neronlux/ntech-team-kit"
  url "https://github.com/neronlux/ntech-team-kit/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "REPLACE_WITH_ACTUAL_SHA256"
  license "MIT"
  head "https://github.com/neronlux/ntech-team-kit.git", branch: "main"

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s -w
      -X main.version=#{version}
      -X github.com/neronlux/ntech-team-kit/internal/kit.defaultKitRoot=#{libexec}
    ]

    system "go", "build",
           *std_go_args(ldflags: ldflags, output: bin/"ntech-team-kit"),
           "./cmd/ntech-team-kit"

    # Install the actual kit contents
    libexec.install "skills", "agents", "commands", "rules", "plugins"
    libexec.install "install.sh", "opencode.jsonc", "AGENTS.md", "package.json", "VERSION"
  end

  def caveats
    <<~EOS
      The kit is now installed, but it is **not** yet active in OpenCode.

      Run the following command to install it into your OpenCode config:

        ntech-team-kit install

      Recommended next steps:

        ntech-team-kit doctor
        ntech-team-kit status

      To enable the background CI watcher plugin:

        echo 'export OPENCODE_NTECH_CI_WATCH=1' >> ~/.zshrc   # or ~/.bashrc

      For more information, see:
        https://github.com/neronlux/ntech-team-kit
    EOS
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/ntech-team-kit version")
    assert_match "ntech-team-kit", shell_output("#{bin}/ntech-team-kit path")
  end
end
