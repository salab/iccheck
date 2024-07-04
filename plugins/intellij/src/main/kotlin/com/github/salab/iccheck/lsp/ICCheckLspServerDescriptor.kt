package com.github.salab.iccheck.lsp

import com.github.salab.iccheck.utils.OsCheck
import com.intellij.execution.configurations.GeneralCommandLine
import com.intellij.openapi.diagnostic.Logger
import com.intellij.openapi.project.Project
import com.intellij.openapi.vfs.VirtualFile
import com.intellij.platform.lsp.api.ProjectWideLspServerDescriptor
import java.io.File
import java.io.FileOutputStream
import java.net.URI
import java.nio.channels.Channels

class ICCheckLspServerDescriptor(project: Project) : ProjectWideLspServerDescriptor(project, "ICCheck") {
    private val LOG = Logger.getInstance(ICCheckLspServerDescriptor::class.java)

    override fun isSupportedFile(file: VirtualFile): Boolean {
        return true
    }

    private fun findExecutableOnPath(name: String): Boolean {
        for (dirname in System.getenv("PATH").split(File.pathSeparator)) {
            val file = File(dirname, name)
            if (file.isFile() && file.canExecute()) {
                return true
            }
        }
        return false
    }

    private fun findExecutableOnWorkdir(name: String): Boolean {
        val workdir = System.getProperty("user.dir")
        val file = File(workdir, name)
        if (file.isFile() && file.canExecute()) {
            return true
        }
        return false
    }

    private fun getDownloadLink(): String {
        val base = "https://github.com/salab/iccheck/releases/download/%s/iccheck_%s_%s_%s"

        val version = "0.3.2" // TODO: refer to config or property?

        val osType = OsCheck.operatingSystemType
        val dlOSName = when (osType) {
            OsCheck.OSType.Windows -> "windows"
            OsCheck.OSType.Linux -> "linux"
            OsCheck.OSType.MacOS -> "darwin"
            else -> throw Exception("OS not supported")
        }

        val dlArchName = System.getProperty("os.arch")

        var url = base.format(version, version, dlOSName, dlArchName)
        if (osType == OsCheck.OSType.Windows) {
            url += ".exe"
        }
        return url
    }

    private fun downloadBinary(url: String, output: String) {
        LOG.info("Downloading ICCheck LSP binary from link %s to %s ...".format(url, output))

        URI(url).toURL().openStream().use { urlSt ->
            Channels.newChannel(urlSt).use { ch ->
                FileOutputStream(output).use { outSt ->
                    val fileCh = outSt.channel
                    fileCh.transferFrom(ch, 0, Long.MAX_VALUE)
                }
            }
        }
    }

    override fun createCommandLine(): GeneralCommandLine {
        // Check PATH
        if (findExecutableOnPath("iccheck")) {
            return GeneralCommandLine("iccheck", "lsp")
        }

        // Check pwd
        if (findExecutableOnWorkdir("iccheck")) {
            return GeneralCommandLine("./iccheck", "lsp")
        }

        // Download from GitHub release
        val url = getDownloadLink()
        val workdir = System.getProperty("user.dir")
        downloadBinary(url, File(workdir, "iccheck").absolutePath)
        return GeneralCommandLine("./iccheck", "lsp")
    }
}
