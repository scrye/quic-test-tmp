import argparse
import itertools
import os
import shutil
import subprocess

OUTPUT_DIR_NAME = "out"

def main():
    parser = argparse.ArgumentParser(description='QUIC test harness.')
    parser.add_argument('--clean', action='store_true')
    parser.add_argument('--size', help='number of KB for transferred files', default=1)
    args = parser.parse_args()

    if args.clean and os.path.exists(OUTPUT_DIR_NAME):
        shutil.rmtree(OUTPUT_DIR_NAME)

    if not os.path.exists(OUTPUT_DIR_NAME):
        os.mkdir(OUTPUT_DIR_NAME)
    else:
        if not os.path.isdir(OUTPUT_DIR_NAME):
            raise ValueError("path is not directory")

    app_path = os.path.join(OUTPUT_DIR_NAME, "main")
    subprocess.run(["go", "build", "-o", app_path, os.path.join("src", "main.go")], check=True)

    goodput = {}
    for mtu in range(1200, 699, -50):
        total_size = run_experiment(app_path, mtu, args.size)
        goodput[mtu] = (int(args.size) * 1024) / total_size
    print(goodput)

def run_experiment(app_path, mtu, size):
    experiment_subdir = os.path.join(OUTPUT_DIR_NAME, f'{mtu:04}')
    os.mkdir(experiment_subdir)

    input_file_path = os.path.join(experiment_subdir, 'input.dat')
    subprocess.run(["dd", "if=/dev/urandom", f'of={input_file_path}', "bs=1024", f'count={size}'], capture_output=True)
    output_file_path = os.path.join(experiment_subdir, 'output.dat')
    stats_file_path = os.path.join(experiment_subdir, 'stats.dat')
    subprocess.run([
        app_path,
        '-input', input_file_path,
        '-output', output_file_path,
        '-stats', stats_file_path,
        '-mtu', str(mtu)],
        check=True)
    subprocess.run(["diff", input_file_path, output_file_path])

    return compute_total_size(stats_file_path)


def compute_total_size(path):
    with open(path) as f:
        message_sizes = (int(s) for s in itertools.islice(f, 1, None))
        return sum(message_sizes)

if __name__ == "__main__":
    main()
