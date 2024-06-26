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

export function activate(context: ExtensionContext) {
	const serverOptions: ServerOptions = {
		command: getBinaryPath(context),
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
