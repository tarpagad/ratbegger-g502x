export namespace main {
	
	export class BackupInfo {
	    path: string;
	    file: string;
	    device: string;
	    model: string;
	    // Go type: time
	    createdAt: any;
	    profiles: number;
	
	    static createFrom(source: any = {}) {
	        return new BackupInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.file = source["file"];
	        this.device = source["device"];
	        this.model = source["model"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.profiles = source["profiles"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MacroEvent {
	    type: number;
	    keycode: number;
	
	    static createFrom(source: any = {}) {
	        return new MacroEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.keycode = source["keycode"];
	    }
	}
	export class Button {
	    path: string;
	    index: number;
	    actionType: number;
	    actionTypes: number[];
	    value: number;
	    macro: MacroEvent[];
	
	    static createFrom(source: any = {}) {
	        return new Button(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.index = source["index"];
	        this.actionType = source["actionType"];
	        this.actionTypes = source["actionTypes"];
	        this.value = source["value"];
	        this.macro = this.convertValues(source["macro"], MacroEvent);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Resolution {
	    path: string;
	    index: number;
	    dpi: number;
	    dpis: number[];
	    active: boolean;
	    default: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Resolution(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.index = source["index"];
	        this.dpi = source["dpi"];
	        this.dpis = source["dpis"];
	        this.active = source["active"];
	        this.default = source["default"];
	    }
	}
	export class Profile {
	    path: string;
	    index: number;
	    name: string;
	    enabled: boolean;
	    active: boolean;
	    reportRate: number;
	    reportRates: number[];
	    resolutions: Resolution[];
	    buttons: Button[];
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.index = source["index"];
	        this.name = source["name"];
	        this.enabled = source["enabled"];
	        this.active = source["active"];
	        this.reportRate = source["reportRate"];
	        this.reportRates = source["reportRates"];
	        this.resolutions = this.convertValues(source["resolutions"], Resolution);
	        this.buttons = this.convertValues(source["buttons"], Button);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Device {
	    path: string;
	    name: string;
	    model: string;
	    profiles: Profile[];
	
	    static createFrom(source: any = {}) {
	        return new Device(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.model = source["model"];
	        this.profiles = this.convertValues(source["profiles"], Profile);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class Status {
	    connected: boolean;
	    apiVersion: number;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.apiVersion = source["apiVersion"];
	        this.error = source["error"];
	    }
	}

}

