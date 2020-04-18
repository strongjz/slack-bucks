.PHONY: get plan apply

all: check clean get plan apply

check:
	terraform -v  >/dev/null 2>&1 || echo "Terraform not installed" || exit 1

clean:
	rm -rf .terraform
	rm -rf plan.out

get:
	terraform get

plan:
	terraform plan -out=plan.out -var-file=secrets.tfvars

apply: plan
	terraform apply -auto-approve -var-file=secrets.tfvars

destroy:
	terraform destroy -var-file=secrets.tfvars