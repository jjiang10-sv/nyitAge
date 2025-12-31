#!/usr/bin/env python3
"""Simple Terraform runner: init/plan/apply/destroy per directory.

Usage examples:
  python main.py --cloud aws --action apply --tfdir terraform/aws --auto-approve
"""
import subprocess
import sys
import os
import click
from dotenv import load_dotenv

load_dotenv()

def run(cmd, cwd=None, env=None):
    print(f"$ {' '.join(cmd)} (cwd={cwd})")
    proc = subprocess.run(cmd, cwd=cwd, env=env)
    if proc.returncode != 0:
        raise SystemExit(proc.returncode)

@click.command()
@click.option("--cloud", type=click.Choice(["aws", "azure"]), required=True, help="Target cloud")
@click.option("--action", type=click.Choice(["init","plan","apply","destroy"]), default="apply")
@click.option("--tfdir", default=None, help="Terraform directory to operate on (overrides cloud default)")
@click.option("--auto-approve", is_flag=True, default=False)
def cli(cloud, action, tfdir, auto_approve):
    if tfdir is None:
        tfdir = os.path.join("terraform", cloud)
    if not os.path.isdir(tfdir):
        print("Terraform dir not found:", tfdir)
        raise SystemExit(1)

    # ensure terraform installed
    try:
        subprocess.run(["terraform", "version"], check=True, stdout=subprocess.DEVNULL)
    except Exception as e:
        print("Terraform CLI not found. Install terraform first.")
        raise SystemExit(1)

    if action == "init":

        run(["terraform", "init", "-input=false"], cwd=tfdir)
    elif action == "plan":
        run(["terraform", "plan", "-input=false", "-out=tfplan"], cwd=tfdir)
    elif action == "apply":
        run(["terraform", "init", "-input=false"], cwd=tfdir)
        plan_path = os.path.join(tfdir, "tfplan")
        if os.path.exists(plan_path):
            if auto_approve:
                run(["terraform", "apply", "-auto-approve", "tfplan"], cwd=tfdir)
            else:
                run(["terraform", "apply", "tfplan"], cwd=tfdir)
        else:
            if auto_approve:
                run(["terraform", "apply", "-auto-approve"], cwd=tfdir)
            else:
                run(["terraform", "apply"], cwd=tfdir)
    elif action == "destroy":
        run(["terraform", "init", "-input=false"], cwd=tfdir)
        cmd = ["terraform", "destroy"] 
        if auto_approve:
            cmd.append("-auto-approve")
        run(cmd, cwd=tfdir)
    else:
        print("unknown action")
        raise SystemExit(1)

if __name__ == "__main__":
    cli()
