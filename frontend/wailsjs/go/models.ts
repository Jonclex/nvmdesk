export namespace main {
	
	export class CurrentInfo {
	    nodeVersion: string;
	    npmVersion: string;
	    nvmVersion: string;
	    nvmRoot: string;
	
	    static createFrom(source: any = {}) {
	        return new CurrentInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodeVersion = source["nodeVersion"];
	        this.npmVersion = source["npmVersion"];
	        this.nvmVersion = source["nvmVersion"];
	        this.nvmRoot = source["nvmRoot"];
	    }
	}
	export class NodeVersion {
	    version: string;
	    isCurrent: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NodeVersion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.isCurrent = source["isCurrent"];
	    }
	}
	export class RemoteVersion {
	    version: string;
	    isLTS: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RemoteVersion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.isLTS = source["isLTS"];
	    }
	}

}

