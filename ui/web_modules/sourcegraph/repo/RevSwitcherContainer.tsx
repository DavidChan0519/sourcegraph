// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Container} from "sourcegraph/Container";
import "sourcegraph/repo/RepoBackend";
import {RepoStore} from "sourcegraph/repo/RepoStore";
import "sourcegraph/tree/TreeBackend";
import {TreeStore} from "sourcegraph/tree/TreeStore";
import {RevSwitcher} from "sourcegraph/repo/RevSwitcher";

interface Props {
	repo: string;
	rev?: string;
	commitID: string;
	repoObj?: any;
	isCloning: boolean;

	// srclibDataVersions is TreeStore.srclibDataVersions.
	srclibDataVersions?: any;

	// to construct URLs
	routes: any[];
	routeParams: any;
}

type State = any;

// RevSwitcherContainer is for standalone RevSwitchers that need to
// be able to respond to changes in RepoStore independently.
export class RevSwitcherContainer extends Container<Props, State> {
	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);
		state.branches = RepoStore.branches;
		state.tags = RepoStore.tags;
		state.srclibDataVersions = TreeStore.srclibDataVersions;
	}

	stores(): FluxUtils.Store<any>[] {
		return [RepoStore, TreeStore];
	}

	render(): JSX.Element | null {
		return (
			<RevSwitcher
				branches={this.state.branches}
				tags={this.state.tags}
				srclibDataVersions={this.state.srclibDataVersions}
				{...this.props} />
			);
	}
}
