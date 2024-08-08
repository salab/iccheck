import { ExtensionContext } from 'vscode';
import {
	LanguageClient,
	LanguageClientOptions,
	ServerOptions
} from 'vscode-languageclient/node';
import * as fs from 'node:fs';

let client: LanguageClient;

const getBinaryPath = (context: ExtensionContext): string => {
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
