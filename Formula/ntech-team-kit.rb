class NtechTeamKit < Formula
  desc "OpenCode and Codex skills, agents, commands, and rules (Cursor Team Kit port)"
  homepage "https://github.com/neronlux/ntech-team-kit"
  url "https://github.com/neronlux/ntech-team-kit/archive/refs/tags/v0.1.32.tar.gz"
  sha256 "9a1fc77e7e955d68c17843ce6dd7dfdba84f6ae8e7f55c18de8591f92d02062a"
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
    libexec.install "opencode.jsonc", "AGENTS.md", "package.json", "VERSION"
  end

  def caveats
    <<~EOS
      The kit contents are installed, but they are **not** yet active in OpenCode.

      Run the following to copy skills, agents, commands and rules into your OpenCode config:

        ntech-team-kit install

      Or install the skills for Codex CLI / IDE / app:

        ntech-team-kit install --target codex

      The CLI is now fully native Go (no shell script delegation), making
      install / update / uninstall / status reliable across platforms.

      Recommended next steps:

        ntech-team-kit doctor
        ntech-team-kit update
        ntech-team-kit status
        ntech-team-kit uninstall     # if you ever want to remove everything

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
