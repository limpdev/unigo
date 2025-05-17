export namespace main {
	
	export class NerdFontIcon {
	    name: string;
	    codepoint: string;
	
	    static createFrom(source: any = {}) {
	        return new NerdFontIcon(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.codepoint = source["codepoint"];
	    }
	}
	export class SvgIcon {
	    name: string;
	    content: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new SvgIcon(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.content = source["content"];
	        this.path = source["path"];
	    }
	}

}

