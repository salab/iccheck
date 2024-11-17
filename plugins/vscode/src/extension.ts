import { ExtensionContext } from 'vscode';
import {
	LanguageClient,
	LanguageClientOptions,
	ServerOptions
} from 'vscode-languageclient/node';
import * as fs from 'node:fs';
import * as path from 'node:path'

let client: LanguageClient;

const findExecutableOnPath = (name: string): string | undefined => {
	for (const dir of process.env.PATH.split(path.delimiter)) {
		const file = path.join(dir, name)
		if (fs.existsSync(file) && !!(fs.statSync(file).mode & fs.constants.S_IXUSR)) {
			return file
		}
	}
	return undefined
}

const getBinaryPath = (context: ExtensionContext): string => {
	// Prefer locally installed binary on PATH
	const binaryPath = findExecutableOnPath('iccheck')
	if (binaryPath) {
		return binaryPath
	}

	// Otherwise, use the bundled binary
	const linux = context.asAbsolutePath('./iccheck')
	if (fs.existsSync(linux)) {
		return linux
	}
	const windows = context.asAbsolutePath('./iccheck.exe')
	if (fs.existsSync(windows)) {
		return windows
	}

	throw new Error('ICCheck binary not found, perhaps an extension release misconfiguration? Please contact the extension developer.')
}

const ensureExecPerm = (path: string) => {
	try {
		fs.accessSync(path, fs.constants.X_OK)
		return
	} catch (_) {
		fs.chmodSync(path, 0o755)
	}
}

export function activate(context: ExtensionContext) {
	const binaryPath = getBinaryPath(context)
	ensureExecPerm(binaryPath)

	const serverOptions: ServerOptions = {
		command: binaryPath,
		args: ['lsp']
	};

	const clientOptions: LanguageClientOptions = {
		documentSelector: [{ scheme: 'file', pattern: '**/*' }]
	};

	client = new LanguageClient(
		'ICCheck',
		'ICCheck',
		serverOptions,
		clientOptions
	);

	client.start();
}

export function deactivate(): Thenable<void> | undefined {
	if (!client) {
		return undefined;
	}
	return client.stop();
}
