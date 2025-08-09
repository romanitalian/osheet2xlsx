class Osheet2xlsx < Formula
  desc "Convert Osheet (.osheet) files to Excel (.xlsx)"
  homepage "https://github.com/romanitalian/osheet2xlsx"
  head "https://github.com/romanitalian/osheet2xlsx.git", branch: "main"

  depends_on "go" => :build

  def install
    commit = Utils.git_head || "unknown"
    build_time = Time.now.utc.strftime("%Y-%m-%dT%H:%M:%SZ")
    ver = version || "HEAD"
    ldflags = [
      "-s -w",
      "-X github.com/romanitalian/osheet2xlsx/v3/cmd.version=#{ver}",
      "-X github.com/romanitalian/osheet2xlsx/v3/cmd.commit=#{commit}",
      "-X github.com/romanitalian/osheet2xlsx/v3/cmd.date=#{build_time}",
    ].join(" ")

    system "go", "build", "-trimpath", "-o", bin/"osheet2xlsx", "-ldflags", ldflags, "."
  end

  test do
    output = shell_output("#{bin}/osheet2xlsx version")
    assert_match "version:", output
  end
end


