from watchfiles import run_process
import subprocess


def run():
    subprocess.run(["python", "main.py"])


if __name__ == "__main__":
    run_process("./", target=run)
