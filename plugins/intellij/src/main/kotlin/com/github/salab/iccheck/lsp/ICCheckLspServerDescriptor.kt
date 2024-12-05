package com.github.salab.iccheck.lsp

import com.github.salab.iccheck.utils.OsCheck
import com.intellij.execution.configurations.GeneralCommandLine
import com.intellij.openapi.diagnostic.Logger
import com.intellij.openapi.project.Project
import com.intellij.openapi.util.io.FileUtilRt
import com.intellij.openapi.vfs.VirtualFile
import com.intellij.platform.lsp.api.ProjectWideLspServerDescriptor
import java.io.BufferedInputStream
import java.io.File
import java.io.FileOutputStream
import java.net.HttpURLConnection
import java.net.URI
import java.nio.file.Files
import java.nio.file.attribute.PosixFilePermission

const val version = "0.7.6" // TODO: refer to config or property?

val tmpFile = FileUtilRt.createTempFile("iccheck-%s".format(version), "")
val dlPath: String = tmpFile.absolutePath

class ICCheckLspServerDescriptor(project: Project) : ProjectWideLspServerDescriptor(project, "ICCheck") {
    private val logger = Logger.getInstance(ICCheckLspServerDescriptor::class.java)

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

    private fun findExecutable(fullPath: String): Boolean {
        val file = File(fullPath)
        return file.isFile() && file.canExecute()
    }

    private fun getDownloadLink(): String {
        val base = "https://github.com/salab/iccheck/releases/download/v%s/iccheck_%s_%s_%s"

        val osType = OsCheck.operatingSystemType
        val dlOSName = when (osType) {
            OsCheck.OSType.Windows -> "windows"
            OsCheck.OSType.Linux -> "linux"
            OsCheck.OSType.MacOS -> "darwin"
            else -> throw Exception("OS not supported")
        }

        val dlArchName = when (val archName = System.getProperty("os.arch")) {
            "amd64", "x64", "x86_64" -> "amd64"
            "arm64", "aarch64" -> "arm64"
            else -> {
                if (archName.startsWith("arm64")) {
                    return "arm64"
                }
                throw Exception("arch name %s not supported".format(archName))
            }
        }

        var url = base.format(version, version, dlOSName, dlArchName)
        if (osType == OsCheck.OSType.Windows) {
            url += ".exe"
        }
        return url
    }

    private fun downloadFile(url: String, output: String) {
        logger.info("Downloading ICCheck LSP binary from link %s to %s ...".format(url, output))

        val httpConn = URI(url).toURL().openConnection() as HttpURLConnection

        httpConn.inputStream.use { inSt ->
            BufferedInputStream(inSt).use { bufInSt ->
                FileOutputStream(output).use { outSt ->
                    val buffer = ByteArray(4096)
                    var bytesRead: Int
                    while ((bufInSt.read(buffer).also { bytesRead = it }) != -1) {
                        outSt.write(buffer, 0, bytesRead)
                    }
                }
            }
        }

        httpConn.disconnect()
    }

    private fun applyExecutePermission(file: String) {
        val perms = HashSet<PosixFilePermission>()
        perms.add(PosixFilePermission.OWNER_READ)
        perms.add(PosixFilePermission.OWNER_WRITE)
        perms.add(PosixFilePermission.OWNER_EXECUTE)
        Files.setPosixFilePermissions(File(file).toPath(), perms)
    }

    override fun createCommandLine(): GeneralCommandLine {
        val basePath = project.basePath

        // Check PATH (local dev)
        if (findExecutableOnPath("iccheck")) {
            logger.info("Selecting LSP binary on PATH")
            return GeneralCommandLine("iccheck", "lsp").withWorkDirectory(basePath)
        }

        // Previously downloaded binaries
        if (findExecutable(dlPath)) {
            logger.info("Selecting LSP binary from previously downloaded file")
            return GeneralCommandLine(dlPath, "lsp").withWorkDirectory(basePath)
        }

        // Download from GitHub release
        val url = getDownloadLink()
        downloadFile(url, dlPath)
        applyExecutePermission(dlPath)

        return GeneralCommandLine(dlPath, "lsp").withWorkDirectory(basePath)
    }
}
