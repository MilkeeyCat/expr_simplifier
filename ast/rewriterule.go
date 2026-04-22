package ast

type RewriteRule struct {
	Pattern Expr
	Result  Expr
}

func (rr *RewriteRule) String() string {
	return rr.Pattern.String() + " => " + rr.Result.String()
}
