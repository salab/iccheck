package com.github.salab.iccheck.utils

import java.util.Locale

/**
 * helper class to check the operating system this Java VM runs in
 *
 * please keep the notes below as a pseudo-license
 *
 * https://stackoverflow.com/questions/228477/how-do-i-programmatically-determine-operating-system-in-java
 * compare to http://svn.terracotta.org/svn/tc/dso/tags/2.6.4/code/base/common/src/com/tc/util/runtime/Os.java
 * http://www.docjar.com/html/api/org/apache/commons/lang/SystemUtils.java.html
 */
object OsCheck {
    // cached result of OS detection
    private var detectedOS: OSType? = null

    val operatingSystemType: OSType?
        /**
         * detect the operating system from the os.name System property and cache
         * the result
         *
         * @returns - the operating system detected
         */
        get() {
            if (detectedOS == null) {
                val OS = System.getProperty("os.name", "generic").lowercase(Locale.ENGLISH)
                detectedOS = if ((OS.indexOf("mac") >= 0) || (OS.indexOf("darwin") >= 0)) {
                    OSType.MacOS
                } else if (OS.indexOf("win") >= 0) {
                    OSType.Windows
                } else if (OS.indexOf("nux") >= 0) {
                    OSType.Linux
                } else {
                    OSType.Other
                }
            }
            return detectedOS
        }

    /**
     * types of Operating Systems
     */
    enum class OSType {
        Windows, MacOS, Linux, Other
    }
}
