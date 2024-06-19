import { ExtensionContext } from 'vscode';
import {
	LanguageClient,
	LanguageClientOptions,
	ServerOptions
} from 'vscode-languageclient/node';

let client: LanguageClient;

export function activate(context: ExtensionContext) {
	const serverOptions: ServerOptions = {
		command: context.asAbsolutePath('../../iccheck'),
		args: ['lsp']
	};

	const clientOptions: LanguageClientOptions = {
		// TODO: select all documents?
		documentSelector: [{ pattern: '**/*.{txt,js,yaml}' }]
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
