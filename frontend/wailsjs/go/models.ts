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
	export class GlobalNpmPackage {
	    name: string;
	    version: string;
	    path: string;
	    sizeBytes: number;
	    sizeLabel: string;
	
	    static createFrom(source: any = {}) {
	        return new GlobalNpmPackage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.path = source["path"];
	        this.sizeBytes = source["sizeBytes"];
	        this.sizeLabel = source["sizeLabel"];
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

