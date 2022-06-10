#!/usr/bin/awk -f

BEGIN {
	FS = ","
	OFS = ","
	print "user_login,user_email,first_name,last_name,user_pass,wp_role,learndash_courses,learndash_groups,display_name,group_leader"
}

NR > 1 {
	printf("%s,%s,%s,%s,,,,%s,%s %s,\n", $3,$3,$1,$2,$6,$1,$2)
}