import matplotlib.pyplot as plt
from matplotlib_set_diagrams import EulerDiagram
import matplotlib_fontja

if __name__ == '__main__':
    # Set up the figure and axis
    fig, ax = plt.subplots(figsize=(7, 4))

    data = {
        (1, 0): 23 - 16,
        (0, 1): 69 - 16,
        (1, 1): 16,
    }
    set_labels = ['Recommendation from ICCheck', 'CBCD Dataset']
    set_colors = ["salmon", "skyblue"]
    diagram = EulerDiagram(data, ax=ax, set_labels=set_labels, set_colors=set_colors)

    for text in diagram.set_label_artists:
        text.set_fontsize(12)
    for key in data:
        text = diagram.subset_label_artists[key]
        text.set_fontsize(12)

    # Add labels, title, and legend
    # ax.set_title('Recommendations in the CBCD Dataset', fontsize=14)

    # Display the plot
    plt.tight_layout()
    # plt.show()
    plt.savefig('3_plot.pdf')
