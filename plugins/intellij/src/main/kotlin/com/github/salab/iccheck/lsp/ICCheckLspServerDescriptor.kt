package com.github.salab.iccheck.lsp

import com.intellij.execution.configurations.GeneralCommandLine
import com.intellij.openapi.project.Project
import com.intellij.openapi.vfs.VirtualFile
import com.intellij.platform.lsp.api.ProjectWideLspServerDescriptor

class ICCheckLspServerDescriptor(project: Project) : ProjectWideLspServerDescriptor(project, "ICCheck") {
    override fun isSupportedFile(file: VirtualFile): Boolean {
        return true
    }

    override fun createCommandLine(): GeneralCommandLine {
        // TODO: how to properly refer to the binary?
        // https://plugins.jetbrains.com/docs/intellij/language-server-protocol.html#integration-overview
        return GeneralCommandLine("./iccheck", "lsp")
    }
}
